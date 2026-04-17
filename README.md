# AI90 - K8s 知识库问答系统

基于 RAG (检索增强生成) 的 Kubernetes 智能问答系统。

## 概述

- **核心功能**：基于 K8s 知识库的智能问答，结合向量检索 + LLM 生成
- **本地 LLM**：使用 Ollama 运行 `qwen2.5:32b` 模型
- **向量存储**：使用 Chroma 存储 K8s 文档向量
- **部署方式**：支持 K8s YAML 部署

## 架构图

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   User Http req │────▶│   AI90 server   │────▶│    Ollama       │
│                 │◀────│                 │◀────│   (LLM)         │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │    Chroma       │
                        │ (Vector Store)  │
                        └─────────────────┘
```

## K8s YAML 部署

### 部署顺序

```bash

# 2. 创建配置
kubectl apply -f k8s/configmap.yaml

# 3. 部署应用
kubectl apply -f k8s/deployment.yaml

```

### 关键配置

- **NodePort**: `31090`
- **镜像**: `ai90:latest`
- **副本数**: 根据需求调整

### 验证命令

```bash
# 查看 Pod 状态
kubectl get pods -l app.kubernetes.io/name=ai90

# 查看服务
kubectl get svc ai90-web

# 查看日志
kubectl logs -l app.kubernetes.io/name=ai90 -f
```

## AI90 主要作用

1. **K8s 知识问答**：基于 K8s 官方文档构建的知识库，回答 K8s 相关问题
2. **向量检索 + LLM 生成**：
   - 使用 Chroma 进行语义相似度检索
   - 使用 Ollama 本地 LLM 生成回答
3. **HTTP API 接口**：提供 RESTful API 供外部调用

## API 格式

### 聊天接口

**POST** `/api/chat`

**Request:**
```json
{
  "question": "什么是 Kubernetes 的 Pod?"
}
```

**Response:**
```json
{
  "answer": "Pod 是 Kubernetes 中最小的部署单元..."
}
```

### 健康检查

**GET** `/health`

**Response:**
```json
{
  "status": "ok"
}
```

## 直接访问 Ollama / Chroma

### Ollama API

```bash
# 生成文本
curl -X POST http://<node>:31134/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen2.5:32b",
    "prompt": "你好"
  }'
```

### Chroma API

```bash
# 查询集合
curl http://<node>:8000/api/v2/tenants/default_tenant/databases/default_database/collections
```

---

**版本**: v1.0  
**维护**: AI90 Team
