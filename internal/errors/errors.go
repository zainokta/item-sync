package errors

import "fmt"

type ErrorCategory int

const (
	CategoryValidation ErrorCategory = iota
	CategoryDatabase
	CategoryCache
	CategoryExternalAPI
	CategoryNotFound
)

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

// Predefined domain errors
func ItemNotFound() *DomainError {
	return &DomainError{
		Code:     "ITEM_NOT_FOUND",
		Message:  "item not found",
		Category: CategoryNotFound,
	}
}

func InvalidItemData(message string) *DomainError {
	return &DomainError{
		Code:     "INVALID_ITEM_DATA",
		Message:  message,
		Category: CategoryValidation,
	}
}

func SyncFailed(cause error) *DomainError {
	return &DomainError{
		Code:     "SYNC_FAILED",
		Message:  "sync failed",
		Category: CategoryExternalAPI,
		Cause:    cause,
	}
}

func CacheFailed(cause error) *DomainError {
	return &DomainError{
		Code:     "CACHE_FAILED",
		Message:  "cache failed",
		Category: CategoryCache,
		Cause:    cause,
	}
}

func ExternalAPIFailed(cause error) *DomainError {
	return &DomainError{
		Code:     "EXTERNAL_API_FAILED",
		Message:  "external API failed",
		Category: CategoryExternalAPI,
		Cause:    cause,
	}
}

func DatabaseError(cause error) *DomainError {
	return &DomainError{
		Code:     "DATABASE_ERROR",
		Message:  "database error",
		Category: CategoryDatabase,
		Cause:    cause,
	}
}

func ItemAlreadyExists() *DomainError {
	return &DomainError{
		Code:     "ITEM_ALREADY_EXISTS",
		Message:  "item already exists",
		Category: CategoryValidation,
	}
}
