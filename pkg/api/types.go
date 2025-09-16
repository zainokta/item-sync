package api

import "github.com/zainokta/item-sync/internal/item/entity"

// PaginationMetadata contains pagination information from external APIs
type PaginationMetadata struct {
	Count    int    `json:"count,omitempty"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
	HasNext  bool   `json:"has_next"`
	HasPrev  bool   `json:"has_prev"`
}

// PaginatedResponse wraps items with pagination metadata
type PaginatedResponse struct {
	Items      []entity.ExternalItem `json:"items"`
	Pagination *PaginationMetadata   `json:"pagination,omitempty"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(items []entity.ExternalItem, pagination *PaginationMetadata) *PaginatedResponse {
	return &PaginatedResponse{
		Items:      items,
		Pagination: pagination,
	}
}

// NewPaginationMetadata creates pagination metadata with hasNext/hasPrev flags
func NewPaginationMetadata(count int, next, previous string) *PaginationMetadata {
	return &PaginationMetadata{
		Count:    count,
		Next:     next,
		Previous: previous,
		HasNext:  next != "",
		HasPrev:  previous != "",
	}
}