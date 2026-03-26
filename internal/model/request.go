package model

// ChatRequest 聊天请求
type ChatRequest struct {
	Question string `json:"question"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Answer string `json:"answer"`
}
