package entity

import (
	"errors"
	"time"
)

type Item struct {
	ID          int                    `json:"id" db:"id"`
	Title       string                 `json:"title" db:"title"`
	Description string                 `json:"description" db:"description"`
	ExternalID  int                    `json:"external_id" db:"external_id"`
	APISource   string                 `json:"api_source" db:"api_source"`
	ExtendInfo  map[string]interface{} `json:"extend_info" db:"extend_info"`
	SyncedAt    time.Time              `json:"synced_at" db:"synced_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type ExternalItem struct {
	ID         int                    `json:"id"`
	Title      string                 `json:"title"`
	ExtendInfo map[string]interface{} `json:"extend_info"`
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

func (i *Item) FromAPIResponse(extItem ExternalItem) {
	i.ExternalID = extItem.ID
	i.Title = extItem.Title
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