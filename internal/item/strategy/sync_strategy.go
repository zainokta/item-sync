package strategy

import (
	"context"

	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/api"
	"github.com/zainokta/item-sync/pkg/logger"
)

// ExternalAPIClient interface for external API calls
type ExternalAPIClient interface {
	Fetch(ctx context.Context, apiName string, operation string, params map[string]interface{}) ([]entity.ExternalItem, error)
	FetchByID(ctx context.Context, apiName string, id int) (entity.ExternalItem, error)
	FetchPaginated(ctx context.Context, apiName string, operation string, params map[string]interface{}) (*api.PaginatedResponse, error)
}

// SyncItemsRequest represents the sync request
type SyncItemsRequest struct {
	ForceSync bool                   `json:"force_sync"`
	APISource string                 `json:"api_source"`
	Operation string                 `json:"operation"`
	Params    map[string]interface{} `json:"params"`
}

type SyncStrategy interface {
	FetchAllItems(ctx context.Context, apiClient ExternalAPIClient, request SyncItemsRequest) ([]entity.ExternalItem, error)
}

func NewSyncStrategy(apiType string, logger logger.Logger) SyncStrategy {
	switch apiType {
	case "pokemon":
		return NewPokemonSyncStrategy(logger)
	case "openweather":
		return NewOpenWeatherSyncStrategy()
	default:
		return &DefaultSyncStrategy{}
	}
}

type DefaultSyncStrategy struct{}

func (d *DefaultSyncStrategy) FetchAllItems(ctx context.Context, apiClient ExternalAPIClient, request SyncItemsRequest) ([]entity.ExternalItem, error) {
	return apiClient.Fetch(ctx, request.APISource, request.Operation, request.Params)
}