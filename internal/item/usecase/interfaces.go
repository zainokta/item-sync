package usecase

import (
	"context"
	"time"

	"github.com/zainokta/item-sync/internal/item/entity"
)

// ItemSaver interface for saving items
type ItemSaver interface {
	Save(ctx context.Context, item entity.Item) error
	Update(ctx context.Context, item entity.Item) error
}

// ItemFinder interface for finding items
type ItemFinder interface {
	FindByID(ctx context.Context, id int) (entity.Item, error)
	FindByExternalID(ctx context.Context, externalID int) (entity.Item, error)
	FindAll(ctx context.Context, limit, offset int) ([]entity.Item, error)
	FindByStatus(ctx context.Context, status string, limit, offset int) ([]entity.Item, error)
	FindByType(ctx context.Context, itemType string, limit, offset int) ([]entity.Item, error)
	FindByAPISource(ctx context.Context, apiSource string, limit, offset int) ([]entity.Item, error)
}

// ItemCache interface for caching
type ItemCache interface {
	GetItems(ctx context.Context, key string) ([]entity.Item, error)
	SetItems(ctx context.Context, key string, items []entity.Item, ttl time.Duration) error
	GetItem(ctx context.Context, key string) (entity.Item, error)
	SetItem(ctx context.Context, key string, item entity.Item, ttl time.Duration) error
	Invalidate(ctx context.Context, key string) error
}

// ExternalAPIClient interface for external API calls
type ExternalAPIClient interface {
	Fetch(ctx context.Context, apiName string, operation string, params map[string]interface{}) ([]entity.ExternalItem, error)
	FetchByID(ctx context.Context, apiName string, id int) (entity.ExternalItem, error)
}

// ItemRepository interface combining saver and finder
type ItemRepository interface {
	ItemSaver
	ItemFinder
}
