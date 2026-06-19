package model

// Role 定义消息中的角色类型。
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message 表示一条 LLM 对话消息，兼容 OpenAI / Claude / 通用格式。
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`

	// ToolCalls 用于 assistant 消息中携带的工具调用请求。
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID 用于 tool 角色消息，对应 assistant 发起调用的 ID。
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ToolCall 表示一次工具调用请求。
type ToolCall struct {
	// ID 是工具调用的唯一标识，由 assistant 生成。
	ID string `json:"id"`

	// Type 固定为 "function"。
	Type string `json:"type"`

	// Function 描述要调用的函数及其参数。
	Function FunctionCall `json:"function"`
}

// FunctionCall 描述函数调用的名称和参数。
type FunctionCall struct {
	// Name 是要调用的函数名。
	Name string `json:"name"`

	// Arguments 是 JSON 格式的参数字符串。
	Arguments string `json:"arguments"`
}

// Tool 定义一个可供 LLM 调用的工具（Function Calling）。
type Tool struct {
	// Type 固定为 "function"。
	Type string `json:"type"`

	// Function 是工具的函数定义。
	Function FunctionDef `json:"function"`
}

// FunctionDef 定义函数工具的元数据和参数格式。
type FunctionDef struct {
	// Name 是函数名，LLM 会通过这个名字发起调用。
	Name string `json:"name"`

	// Description 是函数描述，帮助 LLM 理解何时使用该工具。
	Description string `json:"description"`

	// Parameters 是 JSON Schema 格式的参数定义。
	Parameters map[string]interface{} `json:"parameters"`
}

// ChatRequest 是统一的 LLM 对话请求结构，
// 同时支持普通 Chat 和 Function Calling。
type ChatRequest struct {
	// Model 指定要使用的模型 ID，例如 "claude-sonnet-4"、"gpt-4o"。
	Model string

	// Messages 是对话历史。
	Messages []Message

	// Temperature 控制生成随机性，范围 0~1。
	Temperature float64

	// MaxTokens 限制最大生成 token 数。
	MaxTokens int

	// Tools 是可供 LLM 选择的工具列表，可选。
	Tools []Tool

	// Stream 标识是否使用流式输出。
	Stream bool
}

// NewChatRequest 创建一个最小化的 ChatRequest，
// 便于调用方快速构建请求。
func NewChatRequest(model string) *ChatRequest {
	return &ChatRequest{
		Model:       model,
		Temperature: 0.7,
		MaxTokens:   512,
		Messages:    make([]Message, 0),
	}
}

// AddMessage 追加一条消息，支持链式调用。
func (r *ChatRequest) AddMessage(role Role, content string) *ChatRequest {
	r.Messages = append(r.Messages, Message{Role: role, Content: content})
	return r
}
