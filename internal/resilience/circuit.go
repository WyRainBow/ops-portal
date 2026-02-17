package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker.
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateHalfOpen
	StateOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	name         string
	mu           sync.Mutex
	state        CircuitBreakerState
	openUntil    time.Time
	nextRetry    time.Time
	failureCount int
	successCount int
	requestCount int

	// Configuration
	maxFailures      int
	resetTimeout     time.Duration
	halfOpenAttempts int
}

// CircuitBreakerConfig configures a circuit breaker.
type CircuitBreakerConfig struct {
	MaxFailures      int           // Consecutive failures to open circuit
	ResetTimeout     time.Duration // How long to stay open before half-open
	HalfOpenAttempts int           // Successful attempts to close circuit in half-open
}

// DefaultCircuitBreakerConfig returns default config.
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxFailures:      5,
		ResetTimeout:     1 * time.Minute,
		HalfOpenAttempts: 3,
	}
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		maxFailures:      config.MaxFailures,
		resetTimeout:     config.ResetTimeout,
		halfOpenAttempts: config.HalfOpenAttempts,
	}
}

// Execute runs the function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitBreakerOpen
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// allowRequest determines if a request should be allowed.
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	// Check if we should transition from open to half-open
	if cb.state == StateOpen && now.After(cb.openUntil) {
		cb.state = StateHalfOpen
		cb.nextRetry = now
		return true
	}

	// Allow requests in closed and half-open states
	return cb.state != StateOpen
}

// recordResult records the result of an execution.
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.requestCount++

	if err == nil {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

// onSuccess handles a successful execution.
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0

	switch cb.state {
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.halfOpenAttempts {
			cb.state = StateClosed
			cb.successCount = 0
		}
	case StateClosed:
		// Already closed, nothing to do
	}
}

// onFailure handles a failed execution.
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.successCount = 0

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.maxFailures {
			cb.openCircuit()
		}
	case StateHalfOpen:
		cb.openCircuit()
	}
}

// openCircuit opens the circuit breaker.
func (cb *CircuitBreaker) openCircuit() {
	cb.state = StateOpen
	cb.openUntil = time.Now().Add(cb.resetTimeout)
}

// State returns the current state.
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Errors
var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// DefaultRetryConfig returns default retry config.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
	}
}

// Retry executes a function with exponential backoff retry.
func Retry(ctx context.Context, config *RetryConfig, fn func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	delay := config.BaseDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if context is cancelled
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return fmt.Errorf("after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithResult executes a function with retry and returns a result.
func RetryWithResult[T any](ctx context.Context, config *RetryConfig, fn func() (T, error)) (T, error) {
	var zero T

	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	delay := config.BaseDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return zero, ctx.Err()
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return zero, err
		}

		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return zero, fmt.Errorf("after %d attempts: %w", config.MaxAttempts, lastErr)
}

// ResilientHTTPClient wraps an HTTP client with resilience features.
type ResilientHTTPClient struct {
	circuitBreaker *CircuitBreaker
	retryConfig    *RetryConfig
}

// NewResilientHTTPClient creates a new resilient HTTP client.
func NewResilientHTTPClient(name string, retryConfig *RetryConfig, cbConfig *CircuitBreakerConfig) *ResilientHTTPClient {
	return &ResilientHTTPClient{
		circuitBreaker: NewCircuitBreaker(name, cbConfig),
		retryConfig:    retryConfig,
	}
}

// Do executes an HTTP request with resilience features.
func (c *ResilientHTTPClient) Do(ctx context.Context, fn func() error) error {
	// Check circuit breaker
	if err := c.circuitBreaker.Execute(func() error {
		// Execute with retry
		return Retry(ctx, c.retryConfig, fn)
	}); err != nil {
		if errors.Is(err, ErrCircuitBreakerOpen) {
			return fmt.Errorf("circuit breaker '%s' is open: %w", c.circuitBreaker.name, err)
		}
		return err
	}

	return nil
}

// State returns the current circuit breaker state.
func (c *ResilientHTTPClient) State() CircuitBreakerState {
	return c.circuitBreaker.State()
}
