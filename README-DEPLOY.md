# AI90 RAG CLI 工具部署指南

## 项目概述

AI90是一个基于RAG（检索增强生成）技术的AI助手命令行工具，专门用于回答Kubernetes相关问题。用户可以通过命令行与AI助手进行交互式问答。

### 核心特性

- 🤖 **智能问答**：基于RAG技术，结合向量检索和LLM生成准确回答
- 💻 **命令行交互**：支持交互式模式和单次提问模式
- ☸️ **云原生部署**：支持Docker和Kubernetes部署
- 🔍 **向量检索**：使用Chroma向量数据库存储和检索知识

### 技术架构

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   User Terminal │────▶│   AI90 CLI      │────▶│    Ollama       │
│   (CLI Tool)    │◀────│   Tool          │◀────│   (LLM)         │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │    Chroma       │
                        │ (Vector Store)  │
                        └─────────────────┘
```

---

## 快速开始

### 方式一：Docker本地运行

```bash
# 1. 构建Docker镜像
docker build -t ai90-cli:latest .

# 2. 运行容器（交互模式）
docker run -it \
  --name ai90-cli \
  -e OLLAMA_URL=http://host.docker.internal:11434 \
  -e CHROMA_URL=http://host.docker.internal:8000 \
  ai90-cli:latest

# 3. 或者单次提问模式
docker run -it \
  --name ai90-cli \
  -e OLLAMA_URL=http://host.docker.internal:11434 \
  -e CHROMA_URL=http://host.docker.internal:8000 \
  ai90-cli:latest -s -q "什么是Kubernetes的Pod?"
```

### 方式二：Kubernetes部署

```bash
# 1. 确保K8s集群已就绪，且Ollama和Chroma已部署

# 2. 应用K8s配置
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml

# 3. 验证部署
kubectl get pods -l app.kubernetes.io/name=ai90

# 4. 进入Pod运行CLI
kubectl exec -it deployment/ai90-cli -- /app/ai90
```

---

## 详细部署指南

### 前置条件

#### 必需组件

| 组件 | 版本要求 | 说明 |
|------|---------|------|
| Docker | 20.10+ | 用于构建和运行容器 |
| Kubernetes | 1.25+ | 用于生产环境部署 |
| kubectl | 1.25+ | K8s命令行工具 |
| Ollama | 最新版 | LLM推理服务，需预装qwen2和bge-m3模型 |
| Chroma | 0.4+ | 向量数据库服务 |

#### 预装模型

在Ollama中安装所需模型：

```bash
# 安装LLM模型
ollama pull qwen2

# 安装Embedding模型
ollama pull bge-m3
```

#### 准备知识库

确保Chroma中已有向量化的知识数据：

```bash
# 运行数据导入工具
go run cmd/ingest/main.go
```

---

### 构建Docker镜像

#### 本地构建

```bash
# 进入项目根目录
cd ai90

# 构建镜像（多阶段构建，自动优化体积）
docker build -t ai90-cli:latest .

# 查看构建结果
docker images | grep ai90-cli
```

#### 多架构构建（可选）

```bash
# 创建buildx构建器
docker buildx create --use

# 构建多平台镜像
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t ai90-cli:latest \
  --push .
```

---

### Kubernetes部署

#### 1. 部署到K8s集群

```bash
# 应用配置
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
```

#### 2. 验证部署状态

```bash
# 查看Pod状态
kubectl get pods -l app.kubernetes.io/name=ai90

# 查看部署详情
kubectl describe deployment ai90-cli

# 查看Pod日志
kubectl logs -f deployment/ai90-cli
```

#### 3. 使用CLI工具

```bash
# 进入Pod运行交互式CLI
kubectl exec -it deployment/ai90-cli -- /app/ai90

# 或者单次提问
kubectl exec -it deployment/ai90-cli -- /app/ai90 -s -q "什么是Pod?"
```

---

## 配置说明

### 环境变量列表

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `OLLAMA_URL` | `http://ollama.default.svc.cluster.local:11434` | Ollama服务地址 |
| `CHROMA_URL` | `http://chroma-service.default.svc.cluster.local:8000` | Chroma服务地址 |
| `COLLECTION_NAME` | `k8s_docs_zh` | 向量集合名称 |
| `MODEL_LLM` | `qwen2` | LLM模型名称 |
| `MODEL_EMBEDDING` | `bge-m3` | Embedding模型名称 |
| `LOG_LEVEL` | `info` | 日志级别 |

### 修改配置

编辑 [`k8s/configmap.yaml`](k8s/configmap.yaml:1)：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ai90-config
data:
  OLLAMA_URL: "http://your-ollama:11434"  # 修改为你的Ollama地址
  CHROMA_URL: "http://your-chroma:8000"   # 修改为你的Chroma地址
  COLLECTION_NAME: "k8s_docs_zh"
  MODEL_LLM: "qwen2"
  MODEL_EMBEDDING: "bge-m3"
```

应用更新：

```bash
kubectl apply -f k8s/configmap.yaml

# 重启Deployment使配置生效
kubectl rollout restart deployment/ai90-cli
```

---

## CLI使用指南

### 交互模式

```bash
# 启动交互式CLI
./ai90

# 或者使用Docker
docker run -it ai90-cli:latest
```

交互模式命令：
- 直接输入问题获取AI回答
- `exit` / `quit` / `q`: 退出程序
- `help` / `h`: 显示帮助信息

### 单次提问模式

```bash
# 使用 -s 和 -q 参数
./ai90 -s -q "什么是Kubernetes的Pod?"

# 或者简写形式
./ai90 -s 什么是Kubernetes的Pod?
```

---

## 故障排查

### 常见问题及解决方法

#### 1. Pod无法启动

**排查步骤：**

```bash
# 查看Pod事件
kubectl describe pod <pod-name>

# 查看容器日志
kubectl logs <pod-name>

# 查看之前的容器日志（如果已重启）
kubectl logs <pod-name> --previous
```

**常见原因：**

- **镜像拉取失败**：检查镜像名称和仓库访问权限
- **配置错误**：检查ConfigMap中的环境变量

#### 2. 无法连接到Ollama

**现象：**

日志显示连接Ollama超时或拒绝连接。

**解决方法：**

```bash
# 检查Ollama服务状态
kubectl get svc ollama -n default

# 测试Ollama连通性
kubectl exec -it deployment/ai90-cli -- \
  wget -qO- http://ollama.default.svc.cluster.local:11434/api/tags

# 如果Ollama在集群外，修改ConfigMap中的OLLAMA_URL
```

#### 3. 无法连接到Chroma

**现象：**

向量检索失败或返回空结果。

**解决方法：**

```bash
# 检查Chroma服务
kubectl get svc chroma-service -n default

# 验证Chroma连通性
kubectl exec -it deployment/ai90-cli -- \
  wget -qO- http://chroma-service.default.svc.cluster.local:8000/api/v1/heartbeat

# 检查集合是否存在
curl "http://chroma-service:8000/api/v1/collections"
```

#### 4. 模型未找到

**现象：**

```
Error: model 'qwen2' not found
```

**解决方法：**

```bash
# 在Ollama所在节点执行
ollama pull qwen2
ollama pull bge-m3

# 验证模型已安装
ollama list
```

---

## 目录结构说明

部署相关文件的组织结构：

```
ai90/
├── README-DEPLOY.md          # 本部署指南
├── Dockerfile                # Docker镜像构建文件
├── .dockerignore             # Docker构建忽略文件
├── go.mod                    # Go模块定义
├── go.sum                    # Go依赖校验
├── cmd/
│   ├── main.go              # CLI工具入口
│   └── ingest/              # 数据导入工具
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go        # 配置管理
│   ├── provider/
│   │   ├── chroma.go        # Chroma客户端
│   │   └── ollama.go        # Ollama客户端
│   └── service/
│       └── ragEngine.go     # RAG业务逻辑
├── k8s/                      # Kubernetes配置
│   ├── configmap.yaml       # 环境变量配置
│   └── deployment.yaml      # 应用部署配置
└── plans/
    └── web-service-architecture.md  # 架构设计文档
```

### 关键文件说明

| 文件 | 用途 |
|------|------|
| [`Dockerfile`](Dockerfile:1) | 多阶段构建，生成轻量级CLI镜像 |
| [`k8s/deployment.yaml`](k8s/deployment.yaml:1) | Deployment配置，定义Pod模板和资源限制 |
| [`k8s/configmap.yaml`](k8s/configmap.yaml:1) | ConfigMap配置，存储环境变量 |
| [`internal/config/config.go`](internal/config/config.go:1) | 配置加载逻辑，支持环境变量覆盖 |

---

## 附录

### 常用命令速查

```bash
# 构建和推送
docker build -t ai90-cli:latest .
docker tag ai90-cli:latest registry/ai90-cli:latest
docker push registry/ai90-cli:latest

# K8s部署
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl delete -f k8s/

# 查看状态
kubectl get pods -l app.kubernetes.io/name=ai90
kubectl logs -f deployment/ai90-cli

# 进入容器调试
kubectl exec -it deployment/ai90-cli -- /bin/sh

# 运行CLI
kubectl exec -it deployment/ai90-cli -- /app/ai90
```

### 资源限制参考

| 场景 | CPU请求 | CPU限制 | 内存请求 | 内存限制 |
|------|---------|---------|----------|----------|
| 开发环境 | 100m | 500m | 128Mi | 512Mi |
| 生产环境 | 200m | 1000m | 256Mi | 1Gi |

---

**文档版本：** 1.0.0  
**最后更新：** 2026-03-24  
**维护者：** AI90 Team
