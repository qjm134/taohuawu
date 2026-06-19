package model

import (
	"context"
)

// Provider 定义 LLM 服务提供者的统一接口。
// 所有具体厂商（Claude、OpenAI 等）都需要实现该接口，
// 以便 router 能够以统一的方式进行模型选择、请求路由和降级。
type Provider interface {
	// Name 返回 provider 的唯一标识，例如 "claude"、"openai"。
	Name() string

	// AvailableModels 返回该 provider 下可用的模型 ID 列表。
	AvailableModels() []string

	// Chat 发起一次非流式对话请求。
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// StreamChat 发起一次 SSE 流式对话请求，通过 channel 返回数据块。
	StreamChat(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error)

	// InputPricePer1K 返回输入 token 每 1K 的价格（美元）。
	InputPricePer1K() float64

	// OutputPricePer1K 返回输出 token 每 1K 的价格（美元）。
	OutputPricePer1K() float64

	// MaxContextLength 返回该 provider 支持的最大上下文长度（token 数）。
	MaxContextLength() int
}
