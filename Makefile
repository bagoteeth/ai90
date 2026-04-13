# Makefile for AI90 Project
# 编译目标：
# - target/ai90: 主服务程序 (cmd/server/main.go)
# - target/ingest-tool: 数据导入工具 (cmd/ingest/main.go)
# - target/ai90.tar: Docker 镜像

# 编译参数
export GOOS := linux
export GOARCH := amd64
export CGO_ENABLED := 0

# 目录
TARGET_DIR := target
CMD_DIR := cmd

# 目标文件
SERVER_BINARY := $(TARGET_DIR)/ai90
INGEST_BINARY := $(TARGET_DIR)/ingest-tool
DOCKER_TAR := $(TARGET_DIR)/ai90.tar

# 源文件
SERVER_SRC := $(CMD_DIR)/server/main.go
INGEST_SRC := $(CMD_DIR)/ingest/main.go
DOCKERFILE := Dockerfile

# 默认目标
.PHONY: all clean build docker

all: clean-all build docker

# 创建目标目录
$(TARGET_DIR):
	mkdir $(TARGET_DIR)
	mkdir $(TARGET_DIR)\skills

# 编译主服务程序
build-server: $(SERVER_BINARY)

$(SERVER_BINARY): $(SERVER_SRC) | $(TARGET_DIR)
	@echo "Building server binary: $(SERVER_BINARY)"
	go build -ldflags="-s -w" -o $(SERVER_BINARY) $(SERVER_SRC)
	cp internal/skill/*.md $(TARGET_DIR)/skills/

	@echo "Server binary built successfully"

# 编译数据导入工具
build-ingest: $(INGEST_BINARY)

$(INGEST_BINARY): $(INGEST_SRC) | $(TARGET_DIR)
	@echo "Building ingest tool binary: $(INGEST_BINARY)"
	go build -ldflags="-s -w" -o $(INGEST_BINARY) $(INGEST_SRC)
	@echo "Ingest tool binary built successfully"

# 编译所有二进制文件
build: build-server build-ingest

# 构建 Docker 镜像并保存为 tar 文件
docker: $(DOCKER_TAR)

$(DOCKER_TAR): $(DOCKERFILE) | $(TARGET_DIR)
	@echo "Building Docker image: ai90:latest"
	docker build -t ai90:latest -f $(DOCKERFILE) .
	@echo "Saving Docker image to: $(DOCKER_TAR)"
	docker save -o $(DOCKER_TAR) ai90:latest
	@echo "Docker image saved successfully"

# 清理所有生成的文件
clean:
	@echo "Cleaning up build artifacts..."
	@if exist $(TARGET_DIR) rmdir /s /q $(TARGET_DIR)
	@echo "Clean complete"

# 清理 Docker 镜像
clean-docker:
	@echo "Removing Docker image ai90:latest..."
	-docker rmi ai90:latest 2>nul
	@echo "Docker cleanup complete"

# 完整清理（包括 Docker）
clean-all: clean clean-docker

# 帮助信息
help:
	@echo "AI90 Makefile - 可用目标:"
	@echo ""
	@echo "  make build          - 编译所有二进制文件 (server + ingest)"
	@echo "  make build-server   - 编译主服务程序 (target/ai90)"
	@echo "  make build-ingest   - 编译数据导入工具 (target/ingest-tool)"
	@echo "  make docker         - 构建 Docker 镜像并保存为 tar 文件"
	@echo "  make all            - 编译所有二进制文件和 Docker 镜像"
	@echo "  make clean          - 清理所有生成的文件"
	@echo "  make clean-docker   - 清理 Docker 镜像"
	@echo "  make clean-all      - 完整清理（包括 Docker）"
	@echo "  make help           - 显示此帮助信息"
	@echo ""
	@echo "编译参数:"
	@echo "  GOOS=$(GOOS), GOARCH=$(GOARCH), CGO_ENABLED=$(CGO_ENABLED)"
