package main

import (
	"log"
	"net/http"
	"os"

	"ai90/internal/config"
	"ai90/internal/handler"
)

func main() {
	// 加载配置
	config.GlobalCfg = *config.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	// 设置路由
	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat", handler.ChatHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 静态文件服务
	mux.Handle("/", http.FileServer(http.Dir("./web/static")))

	log.Printf("Server starting on port %s", port)
	log.Printf("Ollama: %s, Chroma: %s", config.GlobalCfg.Ollama.URL, config.GlobalCfg.Chroma.URL)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
