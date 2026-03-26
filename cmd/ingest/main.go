package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// --- 配置参数 ---
const (
	OllamaURL        = "http://127.0.0.1:11434"
	ChromaURL        = "http://127.0.0.1:8000"
	EmbeddingModel   = "bge-m3"
	CollectionName   = "k8s_docs_zh"
	DocRoot          = "/root/testai/kubernetes.io/zh-cn"
	ChunkSize        = 500 // 每个chunk字符数
	MaxChunksPerFile = 30  // 单文件最大chunk数（熔断阈值）
	MaxTotalChunks   = 200 // 全局总chunk数上限（200*500=10万字）
)

// 全局原子计数器，保证并发安全（即使单协程也用原子操作，避免竞态）
var totalChunks uint32

func main() {
	fmt.Println("🚀 启动 K8s 文档入库工具 (V2 协议 + 性能优化版)...")
	fmt.Printf("🔧 配置：单chunk %d字符 | 单文件最大%d块 | 全局最大%d块\n", ChunkSize, MaxChunksPerFile, MaxTotalChunks)

	// 1. 初始化 Collection (使用 V2 接口)
	colID, err := getOrCreateCollectionV2(CollectionName)
	if err != nil {
		log.Fatalf("❌ 无法连接 Chroma: %v \n请确保执行了: kubectl port-forward svc/chroma-service 8000:8000", err)
	}
	fmt.Printf("✅ 向量库就绪: %s\n", colID)

	// 2. 遍历扫描
	fileCount := 0
	err = filepath.Walk(DocRoot, func(path string, info os.FileInfo, err error) error {
		// 检查全局chunk是否已达上限，达到则终止遍历
		if atomic.LoadUint32(&totalChunks) >= MaxTotalChunks {
			fmt.Printf("\n🛑 全局chunk数已达上限(%d)，停止遍历\n", MaxTotalChunks)
			return filepath.SkipAll // 终止整个遍历
		}

		if err != nil {
			log.Printf("⚠️  文件访问失败: %s, 错误: %v", path, err)
			return nil // 跳过错误文件，继续遍历
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".md") {
			processFile(path, colID)
			fileCount++
		}
		return nil
	})

	if err != nil {
		log.Fatalf("❌ 遍历失败: %v", err)
	}
	fmt.Printf("\n🎉 任务结束！共处理文件: %d | 总入库chunk数: %d\n", fileCount, atomic.LoadUint32(&totalChunks))
}

func processFile(path string, colID string) {
	// 提前检查全局上限，避免无效处理
	if atomic.LoadUint32(&totalChunks) >= MaxTotalChunks {
		return
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("⚠️  读取文件失败: %s, 错误: %v", path, err)
		return
	}

	var text string
	if strings.HasSuffix(path, ".html") {
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
		if err != nil {
			log.Printf("⚠️  解析HTML失败: %s, 错误: %v", path, err)
			return
		}
		// 精准提取：只拿正文，过滤掉侧边栏和页脚
		text = doc.Find("main, .td-content, article").Text()
	} else {
		text = string(content)
	}

	text = strings.TrimSpace(text)
	// 调整过滤阈值日志，便于排查
	if len(text) < 150 {
		log.Printf("⚠️  文件内容过短跳过: %s (长度: %d)", path, len(text))
		return
	}

	// 3. 执行切片并判断长度
	chunks := splitText(text, ChunkSize)

	// 核心改动：异常文件跳过逻辑
	if len(chunks) > MaxChunksPerFile {
		fmt.Printf("⚠️  跳过超长文件 (疑似索引页): %s [%d 块]\n", path, len(chunks))
		return
	}

	fmt.Printf("📂 正在处理: %s [%d 块] | 当前全局chunk数: %d\n", path, len(chunks), atomic.LoadUint32(&totalChunks))

	// 计算当前文件能处理的chunk数（避免超全局上限）
	remaining := MaxTotalChunks - atomic.LoadUint32(&totalChunks)
	var processNum int
	if uint32(len(chunks)) <= remaining {
		processNum = len(chunks)
	} else {
		processNum = int(remaining)
		fmt.Printf("⚠️  全局chunk即将达上限，当前文件仅处理前%d块\n", processNum)
	}

	if processNum <= 0 {
		return
	}

	// 处理当前文件的有效chunk
	for i := 0; i < processNum; i++ {
		chunk := chunks[i]
		vec, err := getEmbedding(chunk)
		if err != nil {
			log.Printf("  ❌ Embedding 失败: %s, 块%d, 错误: %v", path, i, err)
			continue
		}

		docID := fmt.Sprintf("%s-%d", path, i)
		err = saveToChromaV2(colID, docID, vec, chunk, path)
		if err != nil {
			log.Printf("  ❌ 写入Chroma失败: %s, 块%d, 错误: %v", path, i, err)
			continue
		}

		// 全局计数+1
		atomic.AddUint32(&totalChunks, 1)
		fmt.Printf("  ✅ 入库成功: %s, 块%d | 累计chunk: %d\n", path, i, atomic.LoadUint32(&totalChunks))

		// 稍微停顿，防止在没有 GPU 的情况下压垮 Ollama
		time.Sleep(10 * time.Millisecond)
	}
}

// --- V2 API 实现函数 ---

func getOrCreateCollectionV2(name string) (string, error) {
	url := fmt.Sprintf("%s/api/v2/tenants/default_tenant/databases/default_database/collections", ChromaURL)

	// 1. 先尝试直接获取列表，看是否存在
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取集合列表失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应码
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("获取集合列表响应异常: 状态码%d, 内容: %s", resp.StatusCode, string(body))
	}

	var collections []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return "", fmt.Errorf("解析集合列表失败: %v", err)
	}

	// 查找已存在的集合
	for _, col := range collections {
		if col.Name == name {
			fmt.Printf("ℹ️  找到已存在的集合: %s (ID: %s)\n", name, col.ID)
			return col.ID, nil
		}
	}

	// 2. 如果没找到，则创建
	payload, _ := json.Marshal(map[string]interface{}{
		"name":          name,
		"get_or_create": true,
	})

	resp, err = http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("创建集合失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查创建响应码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("创建集合响应异常: 状态码%d, 内容: %s", resp.StatusCode, string(body))
	}

	// 打印原始 Body 用于调试
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("ℹ️  创建集合响应: %s\n", string(respBody))

	var res struct {
		ID string `json:"id"`
	}
	// 兼容某些版本可能返回数组的情况
	if err := json.Unmarshal(respBody, &res); err != nil || res.ID == "" {
		var resList []struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(respBody, &resList); err == nil && len(resList) > 0 {
			res.ID = resList[0].ID
		} else {
			return "", fmt.Errorf("解析创建集合响应失败: %v, 响应内容: %s", err, string(respBody))
		}
	}

	fmt.Printf("✅ 创建新集合成功: %s (ID: %s)\n", name, res.ID)
	return res.ID, nil
}

// 修复Chroma V2的add接口路径，增加错误校验
func saveToChromaV2(colID, id string, vec []float32, doc, src string) error {
	// V2 正确路径：包含租户/数据库层级
	url := fmt.Sprintf("%s/api/v2/tenants/default_tenant/databases/default_database/collections/%s/add", ChromaURL, colID)
	payload := map[string]interface{}{
		"ids":        []string{id},
		"embeddings": [][]float32{vec},
		"documents":  []string{doc},
		"metadatas":  []map[string]string{{"source": src}},
		"include":    []string{"documents", "metadatas"}, // 显式指定返回字段，兼容部分版本
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化payload失败: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("POST请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("响应异常: 状态码%d, 内容: %s", resp.StatusCode, string(body))
	}

	return nil
}

func getEmbedding(prompt string) ([]float32, error) {
	data, err := json.Marshal(map[string]string{
		"model":  EmbeddingModel,
		"prompt": prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("序列化embedding请求失败: %v", err)
	}

	resp, err := http.Post(OllamaURL+"/api/embeddings", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("请求Ollama失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama响应异常: 状态码%d, 内容: %s", resp.StatusCode, string(body))
	}

	var res struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("解析embedding响应失败: %v", err)
	}
	if len(res.Embedding) == 0 {
		return nil, fmt.Errorf("embedding结果为空")
	}

	return res.Embedding, nil
}

func splitText(t string, size int) []string {
	var chunks []string
	runes := []rune(t)
	// 空文本直接返回空切片
	if len(runes) == 0 {
		return chunks
	}
	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[i:end])
		// 过滤纯空白的chunk
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}
