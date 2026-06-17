package llm

import (
	"context"
)

// Message LLM 消息
type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// LLMRequest LLM 请求
type LLMRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// LLMResponse LLM 响应
type LLMResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// Adapter LLM 适配器接口
type Adapter interface {
	Chat(ctx context.Context, req *LLMRequest) (*LLMResponse, error)
	IsHealthy() bool
}
