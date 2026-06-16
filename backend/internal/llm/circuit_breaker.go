package llm

import (
	"sync"
	"time"
)

// CircuitState 熔断状态
type CircuitState int

const (
	CircuitClosed CircuitState = iota // 关闭（正常）
	CircuitOpen                       // 开启（熔断）
	CircuitHalfOpen                   // 半开启（探测）
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	maxFailures   int
	failureWindow time.Duration
	recoveryTime  time.Duration
	halfOpenLimit int

	mu           sync.RWMutex
	state        CircuitState
	failures     []time.Time
	successCount int
	lastStateChange time.Time
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(maxFailures int, failureWindow, recoveryTime time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:   maxFailures,
		failureWindow: failureWindow,
		recoveryTime:  recoveryTime,
		halfOpenLimit: 3,
		state:        CircuitClosed,
		failures:     make([]time.Time, 0, maxFailures),
	}
}

// Allow 是否允许请求
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	// 清理过期的失败记录
	cb.cleanupFailures(now)

	// 检查是否可以从半开启切换到关闭
	if cb.state == CircuitHalfOpen && cb.successCount >= cb.halfOpenLimit {
		cb.setState(CircuitClosed, now)
	}

	// 检查是否可以从开启切换到半开启
	if cb.state == CircuitOpen {
		if now.Sub(cb.lastStateChange) >= cb.recoveryTime {
			cb.setState(CircuitHalfOpen, now)
		} else {
			return false
		}
	}

	return true
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitHalfOpen {
		cb.successCount++
	} else {
		cb.failures = cb.failures[:0]
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = append(cb.failures, time.Now())

	if cb.state == CircuitHalfOpen {
		cb.setState(CircuitOpen, time.Now())
		cb.successCount = 0
	} else if len(cb.failures) >= cb.maxFailures {
		cb.setState(CircuitOpen, time.Now())
	}
}

// State 获取当前状态
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state CircuitState, now time.Time) {
	cb.state = state
	cb.lastStateChange = now
	if state == CircuitClosed {
		cb.failures = cb.failures[:0]
		cb.successCount = 0
	}
}

// cleanupFailures 清理过期的失败记录
func (cb *CircuitBreaker) cleanupFailures(now time.Time) {
	if len(cb.failures) == 0 {
		return
	}

	validStart := 0
	for i, t := range cb.failures {
		if now.Sub(t) <= cb.failureWindow {
			validStart = i
			break
		}
	}

	if validStart > 0 {
		cb.failures = cb.failures[validStart:]
	}
}