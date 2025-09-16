package dto

import (
	"github.com/zainokta/item-sync/internal/item/entity"
)

// SyncItemsResponse represents the response from syncing items
type SyncItemsResponse struct {
	SuccessCount int            `json:"success_count" example:"15" description:"Number of items successfully synced"`
	FailedCount  int            `json:"failed_count" example:"2" description:"Number of items that failed to sync"`
	Items        []entity.Item  `json:"items,omitempty" description:"List of synced items (optional)"`
	Errors       []string       `json:"errors,omitempty" description:"List of error messages for failed items"`
}

// GetItemsResponse represents the response from listing items
type GetItemsResponse struct {
	Items []entity.Item `json:"items" description:"List of items"`
	Total int            `json:"total" example:"150" description:"Total number of items matching the query"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string      `json:"code" example:"VALIDATION_ERROR" description:"Error code"`
	Message string      `json:"message" example:"Validation failed" description:"Human readable error message"`
	Details interface{} `json:"details,omitempty" description:"Additional error details"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string `json:"status" example:"healthy" description:"Service health status"`
	Timestamp string `json:"timestamp" example:"2024-01-15T10:30:00Z" description:"Current timestamp"`
	Version   string `json:"version" example:"1.0.0" description:"Service version"`
}
