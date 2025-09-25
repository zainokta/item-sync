package entity

import (
	"errors"
	"time"
)

// Item represents a synchronized item from external APIs
type Item struct {
	ID          int                    `json:"id,omitempty" db:"id" example:"1" description:"Internal database ID"`
	Title       string                 `json:"title" db:"title" example:"Pikachu" description:"Item title/name"`
	Description string                 `json:"description" db:"description" example:"Electric-type Pokemon" description:"Item description"`
	ExternalID  int                    `json:"external_id" db:"external_id" example:"25" description:"ID from external API"`
	APISource   string                 `json:"api_source" db:"api_source" example:"pokemon" description:"Source API (pokemon, openweather)"`
	ExtendInfo  map[string]interface{} `json:"extend_info" db:"extend_info" description:"Additional data from external API"`
	SyncedAt    time.Time              `json:"synced_at" db:"last_synced_at" example:"2024-01-15T10:30:00Z" description:"Last sync timestamp"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at" example:"2024-01-15T10:00:00Z" description:"Creation timestamp"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at" example:"2024-01-15T10:30:00Z" description:"Last update timestamp"`
}

// ExternalItem represents an item from external API before transformation
type ExternalItem struct {
	ID         int                    `json:"id" example:"25" description:"External API item ID"`
	Title      string                 `json:"title" example:"Pikachu" description:"External API item title"`
	ExtendInfo map[string]interface{} `json:"extend_info" description:"Raw data from external API"`
}

func (i *Item) Validate() error {
	if i.Title == "" {
		return errors.New("title is required")
	}
	if i.ExternalID <= 0 {
		return errors.New("external_id must be positive")
	}

	return nil
}

func (i *Item) FromAPIResponse(apiSource string, extItem ExternalItem) {
	i.ExternalID = extItem.ID
	i.Title = extItem.Title
	i.APISource = apiSource
	i.ExtendInfo = extItem.ExtendInfo
	i.SyncedAt = time.Now()
}

func NewItem() Item {
	now := time.Now()
	return Item{
		CreatedAt: now,
		UpdatedAt: now,
	}
}
