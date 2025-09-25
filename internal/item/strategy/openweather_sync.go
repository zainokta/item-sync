package strategy

import (
	"context"

	"github.com/zainokta/item-sync/internal/item/entity"
)

type OpenWeatherSyncStrategy struct {
	apiClient ExternalAPIClient
}

func NewOpenWeatherSyncStrategy(apiClient ExternalAPIClient) *OpenWeatherSyncStrategy {
	return &OpenWeatherSyncStrategy{
		apiClient: apiClient,
	}
}

func (o *OpenWeatherSyncStrategy) FetchAllItems(ctx context.Context, request SyncItemsRequest) ([]entity.ExternalItem, error) {
	return o.apiClient.Fetch(ctx, request.APISource, request.Operation, request.Params)
}

func (o *OpenWeatherSyncStrategy) Fetch(ctx context.Context, request SyncItemsRequest) ([]entity.ExternalItem, error) {
	return o.apiClient.Fetch(ctx, request.APISource, request.Operation, request.Params)
}
