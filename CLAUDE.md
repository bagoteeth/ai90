# AI90 - K8s 智能问答系统

## 项目信息

| 属性 | 值 |
|------|-----|
| **名称** | AI90 |
| **作者** | rfjiang |
| **类型** | RAG (检索增强生成) 智能问答系统 |
| **领域** | Kubernetes 知识库问答 |
| **语言** | Go 1.25 |

## 技术架构

### 核心组件

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   User HTTP req │────▶│   AI90 Server   │────▶│    Ollama       │
│                 │◀────│   (Web API)     │◀────│   (LLM)         │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │    Chroma       │
                        │ (Vector Store)  │
                        └─────────────────┘
```

### 技术栈

| 组件 | 技术 | 用途 |
|------|------|------|
| LLM | Ollama (qwen2.5:32b) | 本地大语言模型生成 |
| 向量数据库 | Chroma | 存储 K8s 文档向量 |
| Web 框架 | Go stdlib (net/http) | HTTP API 服务 |
| 部署 | Docker + Kubernetes | 容器化部署 |

## 项目结构

```
ai90/
├── cmd/
│   ├── main.go              # CLI 交互模式 (deprecated)
│   ├── server/main.go       # HTTP Web 服务入口
│   └── ingest/main.go       # 数据导入工具
├── internal/
│   ├── config/              # 配置管理
│   │   └── config.go        # 环境变量配置
│   ├── handler/
│   │   └── chat_handler.go  # HTTP 处理器
│   ├── model/
│   │   └── request.go       # 请求/响应模型
│   ├── provider/
│   │   ├── chroma.go        # Chroma 向量库客户端
│   │   └── ollama.go        # Ollama LLM 客户端
│   └── service/
│       └── ragEngine.go     # RAG 核心逻辑
├── k8s/
│   ├── deployment.yaml      # K8s Deployment + Service
│   ├── configmap.yaml       # K8s ConfigMap
│   └── namespace.yaml       # K8s Namespace
├── target/                  # 构建输出目录
├── web/static/              # 静态文件
├── Dockerfile               # Docker 镜像构建
├── Makefile                 # 构建脚本
├── go.mod                   # Go 依赖
└── README.md                # 详细文档
```

## 核心功能

### 1. RAG 流程 (ExecuteRAG)

```go
1. 获取用户问题的向量嵌入 (Ollama Embedding)
2. 从 Chroma 检索相关知识文档
3. 构造 Prompt (背景知识 + 问题)
4. 调用 LLM 生成回答
```

### 2. HTTP API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/chat` | POST | 聊天问答接口 |
| `/health` | GET | 健康检查 |
| `/` | GET | 静态文件服务 |

**请求示例:**
```bash
curl -X POST http://localhost:8090/api/chat \
  -H "Content-Type: application/json" \
  -d '{"question": "什么是 Kubernetes 的 Pod?"}'
```

## 构建与部署

### Makefile 命令

```bash
make build          # 编译所有二进制文件 (server + ingest)
make build-server   # 编译主服务程序
make build-ingest   # 编译数据导入工具
make docker         # 构建 Docker 镜像
make all            # 编译 + Docker 构建
make clean          # 清理构建产物
make help           # 显示帮助
```

### Docker 构建

```bash
# 构建镜像
docker build -t ai90:latest .

# 保存镜像
docker save -o target/ai90.tar ai90:latest
```

### Kubernetes 部署

```bash
# 部署到 K8s
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml

# 验证
kubectl get pods -l app.kubernetes.io/name=ai90
kubectl get svc ai90-web
```

**NodePort**: `31090`

## 配置说明

### 环境变量

| 变量名 | 默认值 | 描述 |
|--------|--------|------|
| `PORT` | `8090` | HTTP 服务端口 |
| `OLLAMA_URL` | `http://ollama.default.svc.cluster.local:11434` | Ollama 服务地址 |
| `CHROMA_URL` | `http://chroma-service.default.svc.cluster.local:8000` | Chroma 服务地址 |
| `COLLECTION_NAME` | `k8s_docs_zh` | 向量集合名称 |
| `MODEL_LLM` | `qwen2.5:32b` | LLM 模型 |
| `MODEL_EMBEDDING` | `bge-m3` | 嵌入模型 |

## 依赖信息

### Go 依赖

```go
// go.mod
module ai90

go 1.25.0

require (
    github.com/PuerkitoBio/goquery v1.12.0    // HTML 解析
    k8s.io/api v0.35.3                       // K8s API
    k8s.io/client-go v0.35.3                 // K8s 客户端
    gopkg.in/yaml.v3 v3.0.1                  // YAML 解析
)
```

### 外部服务依赖

- **Ollama**: 本地 LLM 服务 (qwen2.5:32b)
- **Chroma**: 向量数据库服务

## 作者与维护

- **作者**: rfjiang
- **项目**: AI90 - K8s 智能问答系统
- **技术栈**: Go + RAG + K8s
