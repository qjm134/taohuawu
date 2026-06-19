package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/watertown/guide/pkg/logging"
	"github.com/watertown/guide/pkg/utils"
)

// GLMAdapter GLM-4.7 适配器
type GLMAdapter struct {
	apiKey     string
	baseURL    string
	model      string
	client     *http.Client
	timeout    time.Duration
	maxRetries int
	retryDelay time.Duration
	circuit    *CircuitBreaker
	logger     logging.Logger
}

// NewGLMAdapter 创建 GLM 适配器
func NewGLMAdapter(apiKey, baseURL, model string, timeout time.Duration, logger logging.Logger) *GLMAdapter {
	return &GLMAdapter{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		client:     &http.Client{Timeout: timeout},
		timeout:    timeout,
		maxRetries: 3,
		retryDelay: 1 * time.Second,
		circuit:    NewCircuitBreaker(5, 30*time.Second, 60*time.Second),
		logger:     logger,
	}
}

// Chat 发送聊天请求
func (a *GLMAdapter) Chat(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	// 检查熔断器
	if !a.circuit.Allow() {
		return nil, fmt.Errorf("circuit breaker is open")
	}

	var response *LLMResponse

	// 带重试
	ctx, cancel := utils.WithTimeoutFrom(ctx, a.timeout)
	defer cancel()

	err := utils.Retry(ctx, func() error {
		body, err := json.Marshal(req)
		if err != nil {
			return err
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL, bytes.NewReader(body))
		if err != nil {
			return err
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

		resp, err := a.client.Do(httpReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("LLM API error: %d - %s", resp.StatusCode, string(data))
		}

		if err := json.Unmarshal(data, &response); err != nil {
			return err
		}

		return nil
	}, utils.WithMaxRetries(a.maxRetries), utils.WithDelay(a.retryDelay))

	if err != nil {
		a.circuit.RecordFailure()
		return nil, err
	}

	a.circuit.RecordSuccess()
	return response, nil
}

// IsHealthy 检查适配器是否健康
func (a *GLMAdapter) IsHealthy() bool {
	return a.circuit.State() == CircuitClosed
}
