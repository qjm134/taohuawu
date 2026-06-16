package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/watertown/guide/internal/cost"
	"github.com/watertown/guide/internal/emotion"
	"github.com/watertown/guide/internal/llm"
	"github.com/watertown/guide/pkg/utils"
)

// Runtime Agent 运行时
type Runtime struct {
	llmAdapter     llm.Adapter
	fallbackAdapter llm.Adapter
	toolRegistry   *ToolRegistry
	sessionManager *SessionManager
	optimizer      *cost.Optimizer
	emotionDetector emotion.Detector
	config         Config
}

// Config Agent 配置
type Config struct {
	MaxRetries     int
	Timeout        time.Duration
	LLMTimeout     time.Duration
	ToolTimeout    time.Duration
}

// NewRuntime 创建运行时
func NewRuntime(
	llmAdapter, fallbackAdapter llm.Adapter,
	toolRegistry *ToolRegistry,
	sessionManager *SessionManager,
	optimizer *cost.Optimizer,
	emotionDetector emotion.Detector,
	config Config,
) *Runtime {
	return &Runtime{
		llmAdapter:      llmAdapter,
		fallbackAdapter: fallbackAdapter,
		toolRegistry:    toolRegistry,
		sessionManager:  sessionManager,
		optimizer:       optimizer,
		emotionDetector: emotionDetector,
		config:          config,
	}
}

// HandleWelcome 处理欢迎
func (r *Runtime) HandleWelcome(ctx context.Context, session *Session) (string, error) {
	ctx, cancel := utils.WithTimeoutFrom(ctx, r.config.Timeout)
	defer cancel()

	// 检查缓存
	if cached, hit := r.optimizer.GetCache("welcome_" + session.PlayerID); hit {
		return cached, nil
	}

	// 构建消息
	prompt := BuildWelcomePrompt(session.Nickname)

	messages := []llm.Message{
		{Role: "system", Content: SystemPrompt},
		{Role: "user", Content: prompt},
	}

	req := &llm.LLMRequest{
		Messages:    messages,
		Model:       "glm-4",
		Temperature: 0.7,
		MaxTokens:   200,
	}

	var response *llm.LLMResponse
	var err error

	// 尝试主 LLM
	if r.llmAdapter.IsHealthy() {
		ctx, cancel := utils.WithTimeoutFrom(ctx, r.config.LLMTimeout)
		response, err = r.llmAdapter.Chat(ctx, req)
		cancel()

		if err != nil {
			// 降级到备用适配器
			response, err = r.fallbackAdapter.Chat(ctx, req)
		}
	} else {
		// 使用兜底回复
		response, err = r.fallbackAdapter.Chat(ctx, req)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get response: %w", err)
	}

	// 提取回复
	reply := response.Choices[0].Message.Content
	session.AddMessage("assistant", reply, "neutral", nil)

	// 缓存回复
	r.optimizer.SetCache("welcome_"+session.PlayerID, reply)

	return reply, nil
}

// HandleChat 处理聊天
func (r *Runtime) HandleChat(ctx context.Context, session *Session, message string) (string, string, error) {
	ctx, cancel := utils.WithTimeoutFrom(ctx, r.config.Timeout)
	defer cancel()

	// 检测情绪
	em := r.emotionDetector.Detect(message)
	emotionStr := string(em)

	// 检查缓存
	if cached, hit := r.optimizer.GetCache(message); hit {
		session.AddMessage("assistant", cached, emotionStr, nil)
		return cached, emotionStr, nil
	}

	// 获取历史消息
	history := session.GetMessages(10)
	messages := []llm.Message{{Role: "system", Content: SystemPrompt}}

	// 添加历史
	for _, msg := range history {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 添加当前消息
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: message,
	})

	req := &llm.LLMRequest{
		Messages:    messages,
		Model:       "glm-4",
		Temperature: 0.7,
		MaxTokens:   300,
	}

	var response *llm.LLMResponse
	var err error

	// 尝试主 LLM
	if r.llmAdapter.IsHealthy() {
		ctx, cancel := utils.WithTimeoutFrom(ctx, r.config.LLMTimeout)
		response, err = r.llmAdapter.Chat(ctx, req)
		cancel()

		if err != nil {
			// 降级到备用适配器
			response, err = r.fallbackAdapter.Chat(ctx, req)
		}
	} else {
		// 使用兜底回复
		response, err = r.fallbackAdapter.Chat(ctx, req)
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to get response: %w", err)
	}

	// 提取回复
	reply := strings.TrimSpace(response.Choices[0].Message.Content)

	// 添加到会话历史
	session.AddMessage("user", message, emotionStr, nil)
	session.AddMessage("assistant", reply, emotionStr, nil)

	// 缓存回复
	r.optimizer.SetCache(message, reply)

	return reply, emotionStr, nil
}

// GetSession 获取会话
func (r *Runtime) GetSession(playerID, tenantID string) *Session {
	return r.sessionManager.GetOrCreate(playerID, tenantID)
}

// MarkVisited 标记已访问
func (r *Runtime) MarkVisited(sessionID string) {
	if session, ok := r.sessionManager.Get(sessionID); ok {
		session.MarkVisited()
	}
}

// UpdateNickname 更新昵称
func (r *Runtime) UpdateNickname(sessionID, nickname string) {
	if session, ok := r.sessionManager.Get(sessionID); ok {
		session.UpdateNickname(nickname)
	}
}