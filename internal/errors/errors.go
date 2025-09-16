package errors

import "fmt"

type ErrorCategory int

type DomainError struct {
	Code     string
	Message  string
	Category ErrorCategory
	Cause    error
	Details  map[string]interface{}
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}

	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

func (e *DomainError) WithDetail(key string, value interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}
