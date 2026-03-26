package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ai90/internal/config"
)

// ChromaQueryRequest 完整的 Chroma V2 查询请求结构
type ChromaQueryRequest struct {
	QueryEmbeddings [][]float32 `json:"query_embeddings"`
	NResults        int         `json:"n_results"`
	Where           interface{} `json:"where,omitempty"`
	WhereDocument   interface{} `json:"where_document,omitempty"`
	Include         []string    `json:"include,omitempty"`
}

// ChromaQueryResponse 完整的 Chroma V2 查询响应结构
type ChromaQueryResponse struct {
	IDs        [][]string                 `json:"ids"`
	Documents  [][]string                 `json:"documents"`
	Metadatas  [][]map[string]interface{} `json:"metadatas"`
	Distances  [][]float32                `json:"distances"`
	Embeddings [][][]float32              `json:"embeddings,omitempty"`
}

// Collection 表示 Chroma 集合
type Collection struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Configuration interface{}            `json:"configuration,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ChromaClient Chroma V2 API 客户端
type ChromaClient struct {
	BaseURL    string
	Tenant     string
	Database   string
	Collection string
	httpClient *http.Client
}

// NewChromaClient 创建新的 Chroma V2 客户端
func NewChromaClient(baseURL, collectionName string) *ChromaClient {
	return &ChromaClient{
		BaseURL:    baseURL,
		Tenant:     "default_tenant",
		Database:   "default_database",
		Collection: collectionName,
		httpClient: &http.Client{},
	}
}

// getCollectionsURL 获取集合列表的 URL
func (c *ChromaClient) getCollectionsURL() string {
	return fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections",
		c.BaseURL, c.Tenant, c.Database)
}

// getCollectionQueryURL 获取集合查询的 URL
func (c *ChromaClient) getCollectionQueryURL(collectionID string) string {
	return fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/query",
		c.BaseURL, c.Tenant, c.Database, collectionID)
}

// getOrCreateCollection 获取或创建集合，返回集合 ID
func (c *ChromaClient) getOrCreateCollection() (string, error) {
	url := c.getCollectionsURL()

	// 1. 先尝试获取现有集合列表
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取集合列表失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("获取集合列表响应异常: 状态码%d, 内容: %s", resp.StatusCode, string(body))
	}

	// 解析集合列表
	var collections []Collection
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return "", fmt.Errorf("解析集合列表失败: %w", err)
	}

	// 查找已存在的集合
	for _, col := range collections {
		if col.Name == c.Collection {
			return col.ID, nil
		}
	}
	return "", fmt.Errorf("no collection %s found\n", c.Collection)

	// 2. 如果集合不存在，则报错
	//payload, _ := json.Marshal(map[string]interface{}{
	//	"name":          c.Collection,
	//	"get_or_create": true,
	//})
	//
	//resp, err = c.httpClient.Post(url, "application/json", bytes.NewBuffer(payload))
	//if err != nil {
	//	return "", fmt.Errorf("创建集合失败: %w", err)
	//}
	//defer resp.Body.Close()
	//
	//// 检查创建响应码
	//if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
	//	body, _ := io.ReadAll(resp.Body)
	//	return "", fmt.Errorf("创建集合响应异常: 状态码%d, 内容: %s", resp.StatusCode, string(body))
	//}
	//
	//// 读取响应体
	//respBody, _ := io.ReadAll(resp.Body)
	//
	//// 解析响应获取集合 ID
	//var res struct {
	//	ID string `json:"id"`
	//}
	//
	//// 兼容某些版本可能返回数组的情况
	//if err := json.Unmarshal(respBody, &res); err != nil || res.ID == "" {
	//	var resList []struct{ ID string `json:"id"` }
	//	if err := json.Unmarshal(respBody, &resList); err == nil && len(resList) > 0 {
	//		res.ID = resList[0].ID
	//	} else {
	//		return "", fmt.Errorf("解析创建集合响应失败: %w, 响应内容: %s", err, string(respBody))
	//	}
	//}
	//
	//return res.ID, nil
}

// Query 执行向量查询
func (c *ChromaClient) Query(collectionID string, vector []float32, nResults int) (*ChromaQueryResponse, error) {
	url := c.getCollectionQueryURL(collectionID)

	reqBody := ChromaQueryRequest{
		QueryEmbeddings: [][]float32{vector},
		NResults:        nResults,
		Include:         []string{"documents", "metadatas", "distances"},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化查询请求失败: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("查询请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("查询响应异常: 状态码%d, 内容: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result ChromaQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析查询响应失败: %w", err)
	}

	return &result, nil
}

// GetRelevantDoc 根据向量查询最相似的文档内容
// 保留原有的函数签名以保持兼容性
func GetRelevantDoc(vector []float32) (string, error) {
	// 创建 Chroma 客户端
	client := NewChromaClient(config.GlobalCfg.Chroma.URL, config.GlobalCfg.Chroma.CollectionName)

	// 获取或创建集合
	collectionID, err := client.getOrCreateCollection()
	if err != nil {
		return "", fmt.Errorf("获取集合失败: %w", err)
	}

	// 执行查询
	result, err := client.Query(collectionID, vector, 1)
	if err != nil {
		return "", fmt.Errorf("查询失败: %w", err)
	}

	// 检查结果并返回文档
	if len(result.Documents) > 0 && len(result.Documents[0]) > 0 {
		return result.Documents[0][0], nil
	}

	return "", fmt.Errorf("未找到相关文档")
}
