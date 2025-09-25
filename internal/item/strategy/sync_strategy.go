package strategy

import (
	"context"
	"errors"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/api"
	"github.com/zainokta/item-sync/pkg/logger"
)

var (
	ErrAPISourceNotSupported = errors.New("api source is not supported yet")
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
	FetchAllItems(ctx context.Context, request SyncItemsRequest) ([]entity.ExternalItem, error)
	Fetch(ctx context.Context, request SyncItemsRequest) ([]entity.ExternalItem, error)
}

func NewSyncStrategy(cfg *config.Config, apiType string, logger logger.Logger) (SyncStrategy, error) {
	switch apiType {
	case "pokemon":
		apiClient, err := api.NewAPIClient(apiType, cfg.API, cfg.Retry, logger)
		if err != nil {
			return nil, err
		}
		return NewPokemonSyncStrategy(logger, apiClient), nil
	case "openweather":
		apiClient, err := api.NewAPIClient(apiType, cfg.API, cfg.Retry, logger)
		if err != nil {
			return nil, err
		}
		return NewOpenWeatherSyncStrategy(apiClient), nil
	default:
		return nil, ErrAPISourceNotSupported
	}
}
