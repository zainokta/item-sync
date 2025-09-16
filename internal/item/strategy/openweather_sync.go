package strategy

import (
	"context"

	"github.com/zainokta/item-sync/internal/item/entity"
)

type OpenWeatherSyncStrategy struct{}

func NewOpenWeatherSyncStrategy() *OpenWeatherSyncStrategy {
	return &OpenWeatherSyncStrategy{}
}

func (o *OpenWeatherSyncStrategy) FetchAllItems(ctx context.Context, apiClient ExternalAPIClient, request SyncItemsRequest) ([]entity.ExternalItem, error) {
	return apiClient.Fetch(ctx, request.APISource, request.Operation, request.Params)
}