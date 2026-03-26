package service

import (
	"ai90/internal/provider"
	"fmt"
)

func ExecuteRAG(userQuery string) (string, error) {
	// 1. 获取用户问题的向量
	vec, err := provider.GetEmbedding(userQuery)
	if err != nil {
		return "", fmt.Errorf("embedding failed: %v", err)
	}

	// 2. 检索知识库（暂时返回模拟或真实文档）
	knowledge, _ := provider.GetRelevantDoc(vec)
	if knowledge == "" {
		knowledge = "未找到相关内部文档。"
	}

	// 3. 构造 Prompt
	prompt := fmt.Sprintf("基于背景知识回答问题。\n背景：%s\n问题：%s", knowledge, userQuery)
	fmt.Println(prompt)

	// 4. 调用生成模型
	return provider.Generate(prompt)
}
