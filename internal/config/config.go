package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	Ollama OllamaConfig
	Chroma ChromaConfig
	Model  ModelConfig
}

// OllamaConfig Ollama配置
type OllamaConfig struct {
	URL string
}

// ChromaConfig Chroma配置
type ChromaConfig struct {
	URL            string
	CollectionName string
}

// ModelConfig 模型配置
type ModelConfig struct {
	LLM       string
	Embedding string
}

var GlobalCfg Config

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		Ollama: OllamaConfig{
			URL: getEnv("OLLAMA_URL", "http://ollama.default.svc.cluster.local:11434"),
		},
		Chroma: ChromaConfig{
			URL:            getEnv("CHROMA_URL", "http://chroma-service.default.svc.cluster.local:8000"),
			CollectionName: getEnv("COLLECTION_NAME", "k8s_docs_zh"),
		},
		Model: ModelConfig{
			LLM:       getEnv("MODEL_LLM", "qwen2.5:32b"),
			Embedding: getEnv("MODEL_EMBEDDING", "bge-m3"),
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 向后兼容的常量定义
const (
	// Ollama 内部地址: <service-name>.<namespace>.svc.cluster.local
	OllamaURL = "http://ollama.default.svc.cluster.local:11434"

	// Chroma 内部地址
	ChromaURL = "http://chroma-service.default.svc.cluster.local:8000"

	// 模型选型
	ModelLLM       = "qwen2.5:32b"
	ModelEmbedding = "bge-m3"

	// 向量库配置
	CollectionName = "k8s_docs_zh"
)
