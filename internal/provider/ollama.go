package provider

import (
	"ai90/internal/config"
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

var client = &http.Client{Timeout: 600 * time.Second}

func GetEmbedding(text string) ([]float32, error) {
	reqBody, _ := json.Marshal(EmbeddingRequest{
		Model:  config.GlobalCfg.Model.Embedding,
		Prompt: text,
	})

	resp, err := client.Post(config.GlobalCfg.Ollama.URL+"/api/embeddings", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.Embedding, nil
}

func Generate(prompt string) (string, error) {
	reqBody, _ := json.Marshal(GenerateRequest{
		Model:  config.GlobalCfg.Model.LLM,
		Prompt: prompt,
		Stream: false,
	})

	resp, err := client.Post(config.GlobalCfg.Ollama.URL+"/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res GenerateResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Response, nil
}
