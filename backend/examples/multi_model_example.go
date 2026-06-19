package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/watertown/guide/internal/llm"
	"github.com/watertown/guide/internal/llm/model"
	"github.com/watertown/guide/internal/llm/providers/claude"
	"github.com/watertown/guide/internal/llm/providers/openai"
	"github.com/watertown/guide/internal/llm/router"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

// ExampleMultiModelRouter 展示如何使用新的多模型路由系统
func ExampleMultiModelRouter() {
	// 1. 创建多模型路由器
	multiRouter := llm.NewMultiModelRouter(logger)

	// 2. 添加 Claude provider
	claudeProvider, err := claude.NewProvider(claude.Config{
		APIKey:           os.Getenv("ANTHROPIC_API_KEY"),
		Model:            "claude-sonnet-4-20250514",
		InputPrice:       0.000003, // $3 per 1M tokens
		OutputPrice:      0.000015, // $15 per 1M tokens
		MaxContextLength: 200000,
	})
	if err != nil {
		fmt.Printf("Failed to create Claude provider: %v\n", err)
		return
	}
	multiRouter.AddProvider(claudeProvider, true)

	// 3. 添加 OpenAI provider
	openaiProvider, err := openai.NewProvider(openai.Config{
		APIKey:           os.Getenv("OPENAI_API_KEY"),
		Model:            "gpt-4o",
		InputPrice:       0.000005, // $5 per 1M tokens
		OutputPrice:      0.000015, // $15 per 1M tokens
		MaxContextLength: 128000,
	})
	if err != nil {
		fmt.Printf("Failed to create OpenAI provider: %v\n", err)
		return
	}
	multiRouter.AddProvider(openaiProvider, true)

	// 4. 配置路由策略：使用降级链
	multiRouter.SetStrategy(router.StrategyFallback)
	multiRouter.SetFallbackChain([]string{"claude", "openai"})

	// 5. 配置能力映射（根据任务类型选择模型）
	capabilities := model.GetProviderCapabilities()
	for taskType, providers := range capabilities {
		multiRouter.SetCapabilityMap(taskType, providers)
	}

	// 6. 获取适配器，集成到现有系统
	adapter := multiRouter.GetAdapter()

	// 7. 使用示例
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &llm.LLMRequest{
		Model:       "auto", // 模型选择由 router 决定
		Temperature: 0.7,
		MaxTokens:   500,
		Messages: []llm.Message{
			{Role: "system", Content: "你是一个友好的助手"},
			{Role: "user", Content: "请解释什么是递归"},
		},
	}

	resp, err := adapter.Chat(ctx, req)
	if err != nil {
		fmt.Printf("Chat failed: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
}

// ExampleCostStrategy 展示基于成本的路由策略
func ExampleCostStrategy() {
	multiRouter := llm.NewMultiModelRouter(logger)

	// 添加低成本模型
	lowCostProvider, _ := claude.NewProvider(claude.Config{
		APIKey:      os.Getenv("ANTHROPIC_API_KEY"),
		Model:       "claude-haiku-3-5",
		InputPrice:  0.00000025,
		OutputPrice: 0.00000125,
	})
	multiRouter.AddProvider(lowCostProvider, true)

	// 添加高质量模型
	highQualityProvider, _ := claude.NewProvider(claude.Config{
		APIKey:      os.Getenv("ANTHROPIC_API_KEY"),
		Model:       "claude-sonnet-4-20250514",
		InputPrice:  0.000003,
		OutputPrice: 0.000015,
	})
	multiRouter.AddProvider(highQualityProvider, true)

	// 使用成本策略：优先选择成本最低的模型
	multiRouter.SetStrategy(router.StrategyCost)

	adapter := multiRouter.GetAdapter()
	// 使用 adapter...
	_ = adapter
}

// ExampleLatencyStrategy 展示基于延迟的路由策略
func ExampleLatencyStrategy() {
	multiRouter := llm.NewMultiModelRouter(logger)

	// 添加快速模型
	fastProvider, _ := openai.NewProvider(openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "gpt-4o-mini",
	})
	multiRouter.AddProvider(fastProvider, true)

	// 添加标准模型
	standardProvider, _ := openai.NewProvider(openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "gpt-4o",
	})
	multiRouter.AddProvider(standardProvider, true)

	// 使用延迟策略：优先选择响应最快的模型
	// EMA 会自动跟踪每个模型的延迟表现
	multiRouter.SetStrategy(router.StrategyLatency)

	adapter := multiRouter.GetAdapter()
	// 使用 adapter...
	_ = adapter
}

// ExampleWeightedStrategy 展示加权路由策略
func ExampleWeightedStrategy() {
	multiRouter := llm.NewMultiModelRouter(logger)

	// 添加模型并设置权重
	claudeProvider, _ := claude.NewProvider(claude.Config{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:  "claude-sonnet-4-20250514",
	})
	multiRouter.AddProvider(claudeProvider, true)
	multiRouter.SetWeight("claude", 0.7) // 70% 流量

	openaiProvider, _ := openai.NewProvider(openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "gpt-4o",
	})
	multiRouter.AddProvider(openaiProvider, true)
	multiRouter.SetWeight("openai", 0.3) // 30% 流量

	// 使用加权策略：按权重随机选择模型
	// 适合 A/B 测试或流量分配
	multiRouter.SetStrategy(router.StrategyWeighted)

	adapter := multiRouter.GetAdapter()
	// 使用 adapter...
	_ = adapter
}

// ExampleCapabilityStrategy 展示基于任务类型的路由策略
func ExampleCapabilityStrategy() {
	multiRouter := llm.NewMultiModelRouter(logger)

	// 添加代码专用模型
	codeProvider, _ := claude.NewProvider(claude.Config{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:  "claude-sonnet-4-20250514",
	})
	multiRouter.AddProvider(codeProvider, true)

	// 添加通用模型
	generalProvider, _ := openai.NewProvider(openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "gpt-4o",
	})
	multiRouter.AddProvider(generalProvider, true)

	// 配置能力映射
	multiRouter.SetCapabilityMap(model.TaskTypeCode, []string{"claude"})
	multiRouter.SetCapabilityMap(model.TaskTypeGeneral, []string{"openai", "claude"})
	multiRouter.SetCapabilityMap(model.TaskTypeReasoning, []string{"claude"})
	multiRouter.SetCapabilityMap(model.TaskTypeChinese, []string{"openai"})

	// 使用能力策略：根据消息内容自动分类任务类型
	// 然后选择最适合该任务的模型
	multiRouter.SetStrategy(router.StrategyCapability)

	adapter := multiRouter.GetAdapter()
	// 使用 adapter...
	_ = adapter
}

// ExampleTaskClassification 展示任务分类功能
func ExampleTaskClassification() {
	// 任务分类是基于关键词的启发式方法
	// 优先级：Code > Reasoning > Chinese > LongText > General

	codeText := "如何优化这个函数的性能？"
	taskType := model.ClassifyTask(codeText)
	fmt.Printf("任务类型: %s\n", taskType) // 输出: code

	reasoningText := "为什么地球是圆的？"
	taskType = model.ClassifyTask(reasoningText)
	fmt.Printf("任务类型: %s\n", taskType) // 输出: reasoning

	chineseText := "今天天气真好，我们去公园散步吧"
	taskType = model.ClassifyTask(chineseText)
	fmt.Printf("任务类型: %s\n", taskType) // 输出: chinese
}

// ExampleTokenEstimation 展示 token 估算功能
func ExampleTokenEstimation() {
	text := "这是一段中文文本"
	tokens := model.EstimateTokens(text)
	fmt.Printf("估算 token 数: %d\n", tokens)

	englishText := "This is an English text for testing"
	tokens = model.EstimateTokens(englishText)
	fmt.Printf("估算 token 数: %d\n", tokens)
}

// ExampleStreamingChat 展示流式聊天功能
func ExampleStreamingChat() {
	multiRouter := llm.NewMultiModelRouter(logger)

	claudeProvider, _ := claude.NewProvider(claude.Config{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:  "claude-sonnet-4-20250514",
	})
	multiRouter.AddProvider(claudeProvider, true)

	adapter := multiRouter.GetRouter()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := &model.ChatRequest{
		Model:       "claude-sonnet-4-20250514",
		Temperature: 0.7,
		MaxTokens:   1000,
		Messages: []model.Message{
			{Role: model.RoleUser, Content: "请写一篇关于人工智能的文章"},
		},
	}

	stream, err := adapter.RouteRequestStream(ctx, req)
	if err != nil {
		fmt.Printf("Stream failed: %v\n", err)
		return
	}

	// 消费流式响应
	fmt.Print("Response: ")
	for chunk := range stream {
		if chunk.Err != nil {
			fmt.Printf("\nError: %v\n", chunk.Err)
			return
		}
		if chunk.Done {
			fmt.Println("\n[Done]")
			return
		}
		fmt.Print(chunk.Content)
	}
}

// ExampleToolCalling 展示工具调用功能
func ExampleToolCalling() {
	multiRouter := llm.NewMultiModelRouter(logger)

	claudeProvider, _ := claude.NewProvider(claude.Config{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:  "claude-sonnet-4-20250514",
	})
	multiRouter.AddProvider(claudeProvider, true)

	adapter := multiRouter.GetAdapter()

	_ = context.Background()

	// 定义天气查询工具
	weatherTool := model.Tool{
		Type: "function",
		Function: model.FunctionDef{
			Name:        "get_weather",
			Description: "获取指定城市的天气信息",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"city": map[string]interface{}{
						"type":        "string",
						"description": "城市名称",
					},
				},
				"required": []string{"city"},
			},
		},
	}

	req := &llm.LLMRequest{
		Model:       "claude-sonnet-4-20250514",
		Temperature: 0.0,
		MaxTokens:   200,
		Messages: []llm.Message{
			{Role: "user", Content: "今天北京的天气怎么样？"},
		},
	}

	// 注意：当前 Adapter 接口不支持 tools，需要扩展
	// 这是演示概念，实际使用需要扩展接口
	_ = req
	_ = weatherTool
	_ = adapter
}

// ExampleModelStats 展示模型统计功能
func ExampleModelStats() {
	stats := model.NewModelStats()

	// 模拟多次请求
	latencies := []time.Duration{
		2 * time.Second,
		3 * time.Second,
		2500 * time.Millisecond,
	}

	for _, lat := range latencies {
		stats.RecordLatency(lat)
	}

	// 记录错误
	stats.RecordError(false)
	stats.RecordError(false)
	stats.RecordError(true)

	// 查看统计
	fmt.Printf("平均延迟: %.2f ms\n", stats.Latency())
	fmt.Printf("错误率: %.2f%%\n", stats.ErrorRate()*100)
	fmt.Printf("综合分数: %.2f\n", stats.Score())
}

func main() {
	// 运行示例
	ExampleMultiModelRouter()
}
