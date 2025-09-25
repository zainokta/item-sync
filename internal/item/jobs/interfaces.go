package jobs

import (
	"context"
	"time"

	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/api"
)

type JobRepository interface {
	CreateSyncJobRecord(ctx context.Context, name string, apiType string) (int64, error)
	UpdateSyncJobRecord(ctx context.Context, jobID int64, status string, processed, succeeded, failed int, lastErr error, executionTime time.Duration) error
}

// ItemSaver interface for saving items
type ItemSaver interface {
	UpsertWithHash(ctx context.Context, apiSource string, externalItem entity.ExternalItem) error
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
	FetchPaginated(ctx context.Context, apiName string, operation string, params map[string]interface{}) (*api.PaginatedResponse, error)
}

