# 多阶段构建 Dockerfile for AI90 Web 服务
# 构建阶段
FROM golang:1.24-alpine AS builder

RUN mkdir -p /app
RUN mkdir -p /app/skills
WORKDIR /app

# 安装CA证书和时区数据（运行时必需）
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# 从构建阶段复制二进制文件
COPY target/ai90 /app/ai90
COPY target/skills/*.md /app/skills/

# 设置文件权限
RUN chmod +x /app/ai90 && \
    chown -R appuser:appgroup /app

# 使用非root用户运行，符合安全最佳实践
USER appuser:appgroup

# 设置环境变量默认值（可在运行时被覆盖）
ENV PORT=8090 \
    OLLAMA_URL=http://ollama.default.svc.cluster.local:11434 \
    CHROMA_URL=http://chroma-service.default.svc.cluster.local:8000 \
    COLLECTION_NAME=k8s_docs_zh \
    MODEL_LLM=qwen2.5:32b \
    MODEL_EMBEDDING=bge-m3 \
    LOG_LEVEL=info

# 暴露端口
EXPOSE 8090

# 入口点 - 启动 Web 服务
ENTRYPOINT ["/app/ai90"]
