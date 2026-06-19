# 多模型配置说明

## 概述

系统采用三层架构管理多模型调用：

1. **Model 层** — 定义 `Provider` 接口、统一的 `ChatRequest/ChatResponse`，支持 Tool Calling 和 SSE 流式
2. **Provider 层** — 具体实现：Claude（`anthropic-sdk-go`）、OpenAI（`go-openai`），以及兼容 OpenAI 格式的 API（GLM、Qwen 等）
3. **Router 层** — 策略引擎，支持 6 种路由策略 + EMA 统计 + 任务分类 + 降级链

## 快速配置

### YAML 配置（推荐）

编辑 `configs/config.yaml`：

```yaml
llm:
  # 模型列表，按优先级排列
  models:
    # Claude Sonnet（首选，代码和推理能力强）
    - name: claude-sonnet-4-20250514
      base_url: ""
      api_key: ${ANTHROPIC_API_KEY}
      enabled: true
      max_tokens: 2000
      temperature: 0.7

    # OpenAI GPT-4o（备选，中文和通用对话好）
    - name: gpt-4o
      base_url: ""
      api_key: ${OPENAI_API_KEY}
      enabled: true
      max_tokens: 2000
      temperature: 0.7

    # 智谱 GLM-4-Flash（低成本兜底）
    - name: glm-4-flash
      base_url: https://open.bigmodel.cn/api/paas/v4/chat/completions
      api_key: ${GLM_API_KEY}
      enabled: true
      max_tokens: 300
      temperature: 0.7

    # 通义千问（可选）
    - name: qwen-turbo
      base_url: https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
      api_key: ${DASHSCOPE_API_KEY}
      enabled: false
      max_tokens: 300
      temperature: 0.7

  # 通用配置
  timeout: 10s        # 请求超时
  max_retries: 3      # 失败重试次数
  retry_delay: 1s     # 重试延迟
  auto_switch: true   # 启用降级链自动切换
```

### Provider 类型自动推断

系统根据模型名称自动识别 Provider 类型：

| 名称模式 | Provider 类型 | 说明 |
|---------|--------------|------|
| 含 `claude` | Claude | 使用 `anthropic-sdk-go` |
| 含 `gpt` / `o1` / `o3` | OpenAI | 使用 `go-openai` |
| 其他 | OpenAI 兼容模式 | GLM、Qwen、DeepSeek 等 |

## 配置参数

### 模型配置

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 模型名称（用于推断 Provider 类型和路由标识） |
| `base_url` | string | 否 | API Base URL（空字符串 = 使用默认端点） |
| `api_key` | string | 是 | API Key，支持 `${ENV_VAR}` 环境变量 |
| `enabled` | bool | 是 | 是否启用 |
| `max_tokens` | int | 否 | 最大生成 token 数，默认 300 |
| `temperature` | float | 否 | 生成温度，默认 0.7 |

### 通用配置

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `timeout` | duration | 否 | 请求超时，默认 10s |
| `max_retries` | int | 否 | 重试次数，默认 3 |
| `retry_delay` | duration | 否 | 重试延迟，默认 1s |
| `auto_switch` | bool | 否 | 是否启用降级自动切换，默认 true |

## 环境变量

### 不同 Provider 的 API Key

```bash
# Claude
export ANTHROPIC_API_KEY="your-anthropic-key"

# OpenAI
export OPENAI_API_KEY="your-openai-key"

# 智谱 GLM
export GLM_API_KEY="your-glm-key"

# 通义千问
export DASHSCOPE_API_KEY="your-dashscope-key"
```

### Windows PowerShell

```powershell
$env:ANTHROPIC_API_KEY = "your-key"
$env:OPENAI_API_KEY = "your-key"
```

## 路由策略

系统支持 6 种路由策略，通过代码配置（`NewMultiModelRouter()`）：

### 1. Fallback（降级链）— 生产推荐

```go
multiRouter.SetStrategy(router.StrategyFallback)
multiRouter.SetFallbackChain([]string{"claude", "openai", "glm-4-flash"})
```

按顺序尝试：主模型 → 通用备选 → 低成本兜底。

**设计**：降级链容忍更高错误率，确保可用性优先于成本。

### 2. Cost（成本优先）

```go
multiRouter.SetStrategy(router.StrategyCost)
```

自动选择输入+输出成本最低的模型。每个模型可配置 `InputPrice` / `OutputPrice`。

**Token 估算**：每 4 字符约 1 token，中文按字节估算。

### 3. Latency（延迟优先）

```go
multiRouter.SetStrategy(router.StrategyLatency)
```

根据 EMA 统计选择延迟最低的模型。

**EMA 算法**：新样本权重 30%，历史权重 70%，平滑异常波动。

### 4. Capability（能力优先）

```go
multiRouter.SetStrategy(router.StrategyCapability)
multiRouter.SetCapabilityMap(model.TaskTypeCode, []string{"claude"})
multiRouter.SetCapabilityMap(model.TaskTypeChinese, []string{"openai"})
multiRouter.SetCapabilityMap(model.TaskTypeReasoning, []string{"claude"})
multiRouter.SetCapabilityMap(model.TaskTypeLongText, []string{"claude"})
multiRouter.SetCapabilityMap(model.TaskTypeGeneral, []string{"openai", "claude"})
```

根据消息内容自动分类任务类型，选择最适合的模型。

**任务分类优先级**：Code > Reasoning > Chinese > LongText > General。

### 5. Weighted（加权）

```go
multiRouter.SetStrategy(router.StrategyWeighted)
multiRouter.SetWeight("claude", 0.7)  // 70% 流量
multiRouter.SetWeight("openai", 0.3)  // 30% 流量
```

按权重随机选择，适合 A/B 测试或灰度发布。

### 6. Fixed（固定）

```go
multiRouter.SetStrategy(router.StrategyFixed)
multiRouter.SetFixedModel("claude")
```

始终使用指定模型，适合开发调试。

## 支持的 Provider

### Claude（原生 SDK）

```yaml
- name: claude-sonnet-4-20250514
  api_key: ${ANTHROPIC_API_KEY}
  enabled: true
  max_tokens: 2000
```

- SDK：`github.com/anthropics/anthropic-sdk-go`
- 支持：Chat / StreamChat / Tool Calling
- 优势：代码生成、长上下文（200K）、推理能力强

### OpenAI（原生 SDK）

```yaml
- name: gpt-4o
  api_key: ${OPENAI_API_KEY}
  enabled: true
  max_tokens: 2000
```

- SDK：`github.com/sashabaranov/go-openai`
- 支持：Chat / StreamChat / Tool Calling
- 优势：中文对话、通用能力强

### 智谱 GLM（OpenAI 兼容模式）

```yaml
- name: glm-4-flash
  base_url: https://open.bigmodel.cn/api/paas/v4/chat/completions
  api_key: ${GLM_API_KEY}
  enabled: true
  max_tokens: 300
```

- 通过 OpenAI 兼容格式接入
- 成本低，适合作为兜底模型

### 通义千问（OpenAI 兼容模式）

```yaml
- name: qwen-turbo
  base_url: https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
  api_key: ${DASHSCOPE_API_KEY}
  enabled: true
  max_tokens: 300
```

- 通过 OpenAI 兼容格式接入

### DeepSeek / 其他 OpenAI 兼容 API

只要支持 OpenAI Chat Completions 格式，都可以通过设置 `base_url` 接入。

## 自动切换机制

### 触发条件

1. **API 请求失败** — 网络错误、超时、API 返回错误
2. **余额不足** — API 返回配额不足
3. **连续失败** — 同一模型连续失败达到 `max_retries` 阈值

### 切换逻辑

```
1. 根据路由策略选择主模型
2. 如果失败：
   → Fallback 策略：按降级链顺序尝试下一个
   → 其他策略：自动降级到 Fallback 链
3. 如果所有模型都失败：使用 FallbackAdapter 返回预设回复
```

### 日志示例

```json
{"level":"info","msg":"Provider registered","name":"claude-sonnet","model":"claude-sonnet-4-20250514","type":"claude"}
{"level":"info","msg":"Provider registered","name":"gpt-4o","model":"gpt-4o","type":"openai"}
{"level":"error","msg":"Chat failed","provider":"claude","error":"rate limit exceeded"}
{"level":"info","msg":"Fallback to next provider","from":"claude","to":"openai"}
```

## EMA 统计详情

Router 为每个 Provider 维护 `ModelStats`，通过 EMA 跟踪运行指标：

```
延迟 EMA：newEMA = 0.3 × currentSample + 0.7 × previousEMA
错误率 EMA：newEMA = 0.3 × (err ? 1.0 : 0.0) + 0.7 × previousEMA
综合评分：Score = Latency + ErrorRate × 10000
```

**设计意图**：
- 错误率以 10000 倍权重放大，确保高错误模型被快速降级
- 30% 新样本权重保证对新数据快速响应，70% 历史权重保证稳定
- 冷启动时给初始延迟 1000ms，避免被过度惩罚

## 代码级配置示例

```go
package main

import (
    "os"
    "github.com/watertown/guide/internal/llm"
    "github.com/watertown/guide/internal/llm/providers/claude"
    "github.com/watertown/guide/internal/llm/providers/openai"
    "github.com/watertown/guide/internal/llm/model"
    "github.com/watertown/guide/internal/llm/router"
)

func main() {
    multiRouter := llm.NewMultiModelRouter()

    // 添加 Claude
    claudeProvider, _ := claude.NewProvider(claude.Config{
        APIKey:           os.Getenv("ANTHROPIC_API_KEY"),
        Model:            "claude-sonnet-4-20250514",
        InputPrice:       0.000003,
        OutputPrice:      0.000015,
        MaxContextLength: 200000,
    })
    multiRouter.AddProvider(claudeProvider, true)

    // 添加 OpenAI
    openaiProvider, _ := openai.NewProvider(openai.Config{
        APIKey:           os.Getenv("OPENAI_API_KEY"),
        Model:            "gpt-4o",
        InputPrice:       0.000005,
        OutputPrice:      0.000015,
        MaxContextLength: 128000,
    })
    multiRouter.AddProvider(openaiProvider, true)

    // 配置降级链
    multiRouter.SetStrategy(router.StrategyFallback)
    multiRouter.SetFallbackChain([]string{"claude", "openai"})

    // 获取适配器（兼容现有系统）
    adapter := multiRouter.GetAdapter()

    // 使用 adapter.Chat(ctx, req) ...
}
```

更多示例见 [`examples/multi_model_example.go`](examples/multi_model_example.go)。

## 常见问题

### Q1: 如何添加新模型？

在 `configs/config.yaml` 的 `llm.models` 中添加：

```yaml
- name: your-model-name
  base_url: https://your-api-endpoint.com/v1/chat/completions
  api_key: ${YOUR_API_KEY}
  enabled: true
  max_tokens: 300
  temperature: 0.7
```

如果模型名称含 `claude` → 自动用 Claude SDK；含 `gpt` → 自动用 OpenAI SDK；其他 → 自动走 OpenAI 兼容格式。

### Q2: 如何切换路由策略？

YAML 配置模式下默认使用 Fallback 策略（按配置顺序降级）。
如需使用其他策略，需要通过代码配置（`NewMultiModelRouter()`）。

### Q3: 如何禁用某个模型？

将 `enabled` 设为 `false`：

```yaml
- name: glm-4
  enabled: false
```

### Q4: 如何调整模型优先级？

调整模型在配置文件中的顺序 — Fallback 策略下，越靠前的模型优先级越高。

### Q5: 所有模型都失败了怎么办？

依次尝试：
1. 降级链中的下一个模型
2. FallbackAdapter 返回预设兜底回复（关键词匹配）

### Q6: 如何配置 Tool Calling？

所有 Provider 都支持 Tool Calling，通过 `model.ChatRequest.Tools` 传入工具定义：

```go
tool := model.Tool{
    Type: "function",
    Function: model.FunctionDef{
        Name:        "get_weather",
        Description: "获取天气",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "city": map[string]any{"type": "string"},
            },
            "required": []string{"city"},
        },
    },
}
```

## 注意事项

1. **API Key 安全**：不要直接写在配置文件中，使用环境变量
2. **Provider 兼容性**：OpenAI 兼容模式下，确保 API 端点支持 `/chat/completions` 格式
3. **成本控制**：不同模型的单价差异很大，建议配置 `InputPrice` / `OutputPrice` 并使用 Cost 策略
4. **测试验证**：添加新模型后先测试，确认 API 格式兼容
5. **流式支持**：所有 Provider 都支持 SSE 流式输出，但当前 `RouterAdapter.Chat()` 使用非流式模式
6. **不使用 LangChain**：仅依赖 `anthropic-sdk-go` 和 `go-openai` 两个轻量 SDK
