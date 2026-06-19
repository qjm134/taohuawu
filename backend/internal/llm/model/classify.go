package model

import (
	"strings"
	"unicode"
)

// TaskType 定义任务类型，用于根据任务特征选择合适的模型。
type TaskType string

const (
	TaskTypeGeneral    TaskType = "general"    // 通用对话
	TaskTypeCode       TaskType = "code"       // 代码相关
	TaskTypeReasoning  TaskType = "reasoning"  // 推理任务
	TaskTypeChinese    TaskType = "chinese"    // 中文内容
	TaskTypeLongText   TaskType = "longtext"   // 长文本处理
)

// ClassifyTask 根据消息内容识别任务类型。
// 使用关键词匹配启发式规则，优先级从高到低：
// 1. Code - 检测代码关键词和符号
// 2. Reasoning - 检测推理类关键词
// 3. Chinese - 检测中文字符比例
// 4. LongText - 检测文本长度
// 5. 默认为 General
func ClassifyTask(text string) TaskType {
	text = strings.ToLower(text)

	// 检测代码任务
	if isCodeTask(text) {
		return TaskTypeCode
	}

	// 检测推理任务
	if isReasoningTask(text) {
		return TaskTypeReasoning
	}

	// 检测中文内容
	if isChineseContent(text) {
		return TaskTypeChinese
	}

	// 检测长文本
	if isLongText(text) {
		return TaskTypeLongText
	}

	return TaskTypeGeneral
}

// isCodeTask 检测是否为代码相关任务。
func isCodeTask(text string) bool {
	codeKeywords := []string{
		"function", "class", "def ", "import ", "from ", "const ", "let ", "var ",
		"interface", "type ", "struct", "enum", "return", "print", "console.log",
		"npm install", "pip install", "git commit", "docker build",
		"bug", "debug", "error", "exception", "stack trace", "compile",
		".js", ".ts", ".py", ".go", ".java", ".cpp", ".rs", ".rb", ".php",
		"<div>", "<script>", "function(", "def ", "class ", "struct ",
		"算法", "数据结构", "代码", "编程", "函数", "类", "对象",
	}

	for _, kw := range codeKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}

	// 检测代码块标记
	if strings.Contains(text, "```") {
		return true
	}

	// 检测大量特殊字符（可能是代码）
	specialCount := 0
	for _, r := range text {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			specialCount++
		}
	}
	if len(text) > 0 && float64(specialCount)/float64(len(text)) > 0.3 {
		return true
	}

	return false
}

// isReasoningTask 检测是否为推理任务。
func isReasoningTask(text string) bool {
	reasoningKeywords := []string{
		"为什么", "how", "why", "explain", "explain", "reason",
		"分析", "比较", "区别", "差异", "关系",
		"assume", "suppose", "given that", "consider", "imply",
		"证明", "推导", "计算", "求解", "逻辑",
		"step", "思考", "推理", "推断", "结论",
		"problem", "solve", "solution", "answer", "question",
		"假设", "如果", "那么", "因此", "所以",
	}

	for _, kw := range reasoningKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

// isChineseContent 检测是否主要为中文内容。
// 如果中文字符占比超过 30%，则认为是中文内容。
func isChineseContent(text string) bool {
	chineseCount := 0
	for _, r := range text {
		if isChineseChar(r) {
			chineseCount++
		}
	}

	if len(text) == 0 {
		return false
	}

	// 中文字符占比超过 30%
	return float64(chineseCount)/float64(len(text)) > 0.3
}

// isChineseChar 判断是否为中文字符。
func isChineseChar(r rune) bool {
	// CJK 统一表意文字范围
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x20000 && r <= 0x2A6DF) ||
		(r >= 0x2A700 && r <= 0x2B73F) ||
		(r >= 0x2B740 && r <= 0x2B81F) ||
		(r >= 0x2B820 && r <= 0x2CEAF) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0x2F800 && r <= 0x2FA1F)
}

// isLongText 检测是否为长文本。
// Token 估算：每4字符约1token，中文按字节估算。
// 超过 2000 tokens 认为是长文本。
func isLongText(text string) bool {
	tokens := EstimateTokens(text)
	return tokens > 2000
}

// GetProviderCapabilities 返回每种任务类型推荐的 provider 类型。
// 这是一个启发式配置，可根据实际需求调整。
func GetProviderCapabilities() map[TaskType][]string {
	return map[TaskType][]string{
		TaskTypeGeneral:   {"claude", "openai"},
		TaskTypeCode:      {"claude", "openai"},
		TaskTypeReasoning: {"claude"},
		TaskTypeChinese:   {"openai"},
		TaskTypeLongText:  {"claude"},
	}
}