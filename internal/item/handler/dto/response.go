package dto

import (
	"github.com/zainokta/item-sync/internal/item/entity"
)

type SyncItemsResponse struct {
	SuccessCount int            `json:"success_count"`
	FailedCount  int            `json:"failed_count"`
	Items        []entity.Item  `json:"items,omitempty"`
	Errors       []string       `json:"errors,omitempty"`
}

type GetItemsResponse struct {
	Items []entity.Item `json:"items"`
	Total int            `json:"total"`
}

type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}
