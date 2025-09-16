package retry

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/pkg/logger"
)

type Retrier struct {
	config config.RetryConfig
	logger logger.Logger
}

func New(config config.RetryConfig, logger logger.Logger) *Retrier {
	return &Retrier{
		config: config,
		logger: logger,
	}
}

func (r *Retrier) Execute(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := r.calculateBackoff(attempt)
			r.logger.Debug("Retrying operation", "attempt", attempt, "delay", delay)
			
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		
		err := operation()
		if err == nil {
			if attempt > 0 {
				r.logger.Info("Operation succeeded after retry", "attempts", attempt+1)
			}
			return nil
		}
		
		lastErr = err
		
		if !r.shouldRetry(err) {
			r.logger.Debug("Operation failed with non-retryable error", "error", err)
			return err
		}
		
		r.logger.Warn("Operation failed, will retry", "attempt", attempt+1, "error", err)
	}
	
	r.logger.Error("Operation failed after all retries", "max_retries", r.config.MaxRetries, "error", lastErr)
	return lastErr
}

func (r *Retrier) ExecuteWithInterface(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	var lastErr error
	
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := r.calculateBackoff(attempt)
			r.logger.Debug("Retrying operation with result", "attempt", attempt, "delay", delay)
			
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return result, ctx.Err()
			}
		}
		
		res, err := operation()
		if err == nil {
			if attempt > 0 {
				r.logger.Info("Operation with result succeeded after retry", "attempts", attempt+1)
			}
			return res, nil
		}
		
		lastErr = err
		
		if !r.shouldRetry(err) {
			r.logger.Debug("Operation with result failed with non-retryable error", "error", err)
			return result, err
		}
		
		r.logger.Warn("Operation with result failed, will retry", "attempt", attempt+1, "error", err)
	}
	
	r.logger.Error("Operation with result failed after all retries", "max_retries", r.config.MaxRetries, "error", lastErr)
	return result, lastErr
}

func (r *Retrier) calculateBackoff(attempt int) time.Duration {
	delay := time.Duration(float64(r.config.InitialDelay) * math.Pow(r.config.BackoffFactor, float64(attempt-1)))
	if delay > r.config.MaxDelay {
		delay = r.config.MaxDelay
	}
	return delay
}

func (r *Retrier) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	
	return true
}

type RetryableError struct {
	Err error
}

func (e RetryableError) Error() string {
	return e.Err.Error()
}

func (e RetryableError) Unwrap() error {
	return e.Err
}

type NonRetryableError struct {
	Err error
}

func (e NonRetryableError) Error() string {
	return e.Err.Error()
}

func (e NonRetryableError) Unwrap() error {
	return e.Err
}

func NewRetryableError(err error) error {
	return RetryableError{Err: err}
}

func NewNonRetryableError(err error) error {
	return NonRetryableError{Err: err}
}

func IsRetryable(err error) bool {
	var retryableErr RetryableError
	var nonRetryableErr NonRetryableError
	
	if errors.As(err, &nonRetryableErr) {
		return false
	}
	
	if errors.As(err, &retryableErr) {
		return true
	}
	
	return true
}