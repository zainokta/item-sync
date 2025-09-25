package circuit

import (
	"errors"
	"sync"
	"time"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/pkg/logger"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	name      string
	threshold int
	timeout   time.Duration
	logger    logger.Logger

	mu            sync.RWMutex
	state         State
	failureCount  int
	successCount  int
	lastFailTime  time.Time
	nextRetryTime time.Time
}

func NewCircuitBreaker(name string, config config.RetryConfig, logger logger.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		name:      name,
		threshold: config.CircuitThreshold,
		timeout:   config.CircuitTimeout,
		logger:    logger,
		state:     StateClosed,
	}
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
	if !cb.canExecute() {
		cb.logger.Warn("Circuit breaker is open, rejecting request", "name", cb.name)
		return ErrCircuitOpen
	}

	err := operation()
	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		return time.Now().After(cb.nextRetryTime)
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.threshold {
			cb.setState(StateOpen)
			cb.nextRetryTime = time.Now().Add(cb.timeout)
			cb.logger.Warn("Circuit breaker opened due to failures",
				"name", cb.name,
				"failures", cb.failureCount,
				"threshold", cb.threshold)
		}
	case StateHalfOpen:
		cb.setState(StateOpen)
		cb.nextRetryTime = time.Now().Add(cb.timeout)
		cb.logger.Warn("Circuit breaker reopened after failed recovery attempt", "name", cb.name)
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.successCount++

	switch cb.state {
	case StateClosed:
		cb.reset()
	case StateHalfOpen:
		cb.setState(StateClosed)
		cb.reset()
		cb.logger.Info("Circuit breaker closed after successful recovery", "name", cb.name)
	case StateOpen:
		if time.Now().After(cb.nextRetryTime) {
			cb.setState(StateHalfOpen)
			cb.logger.Info("Circuit breaker moved to half-open state", "name", cb.name)
		}
	}
}

func (cb *CircuitBreaker) setState(state State) {
	oldState := cb.state
	cb.state = state

	if oldState != state {
		cb.logger.Debug("Circuit breaker state changed",
			"name", cb.name,
			"from", oldState.String(),
			"to", state.String())
	}
}

func (cb *CircuitBreaker) reset() {
	cb.failureCount = 0
	cb.successCount = 0
}

type BreakerManager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
	config   config.RetryConfig
	logger   logger.Logger
}

func NewBreakerManager(config config.RetryConfig, logger logger.Logger) *BreakerManager {
	return &BreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		config:   config,
		logger:   logger,
	}
}

func (bm *BreakerManager) GetBreaker(name string) *CircuitBreaker {
	bm.mu.RLock()
	if breaker, exists := bm.breakers[name]; exists {
		bm.mu.RUnlock()
		return breaker
	}
	bm.mu.RUnlock()

	bm.mu.Lock()
	defer bm.mu.Unlock()

	if breaker, exists := bm.breakers[name]; exists {
		return breaker
	}

	breaker := NewCircuitBreaker(name, bm.config, bm.logger)
	bm.breakers[name] = breaker
	bm.logger.Info("Created new circuit breaker", "name", name)

	return breaker
}
