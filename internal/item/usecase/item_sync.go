package usecase

import (
	"context"

	pkgErrors "github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/internal/item/strategy"
	"github.com/zainokta/item-sync/pkg/logger"
)

type SyncItemsUseCase struct {
	itemRepo  ItemRepository
	apiClient ExternalAPIClient
	cache     ItemCache
	logger    logger.Logger
	strategy  strategy.SyncStrategy
}

func NewSyncItemsUseCase(itemRepo ItemRepository, apiClient ExternalAPIClient, cache ItemCache, logger logger.Logger, apiType string) *SyncItemsUseCase {
	syncStrategy := strategy.NewSyncStrategy(apiType, logger)
	return &SyncItemsUseCase{
		itemRepo:  itemRepo,
		apiClient: apiClient,
		cache:     cache,
		logger:    logger,
		strategy:  syncStrategy,
	}
}

type SyncItemsRequest struct {
	ForceSync bool                   `json:"force_sync"`
	APISource string                 `json:"api_source"`
	Operation string                 `json:"operation"`
	Params    map[string]interface{} `json:"params"`
}

type SyncItemsResponse struct {
	SuccessCount int           `json:"success_count"`
	FailedCount  int           `json:"failed_count"`
	Items        []entity.Item `json:"items,omitempty"`
	Errors       []string      `json:"errors,omitempty"`
}

func (uc *SyncItemsUseCase) Execute(ctx context.Context, req SyncItemsRequest) (SyncItemsResponse, error) {
	strategyReq := strategy.SyncItemsRequest{
		ForceSync: req.ForceSync,
		APISource: req.APISource,
		Operation: req.Operation,
		Params:    req.Params,
	}

	externalItems, err := uc.strategy.FetchAllItems(ctx, uc.apiClient, strategyReq)
	if err != nil {
		return SyncItemsResponse{}, pkgErrors.ExternalAPIFailed(err)
	}

	response := SyncItemsResponse{
		SuccessCount: 0,
		FailedCount:  0,
		Errors:       make([]string, 0),
		Items:        make([]entity.Item, 0),
	}

	for _, externalItem := range externalItems {
		if err := uc.itemRepo.UpsertWithHash(ctx, req.APISource, externalItem); err != nil {
			uc.logger.Error("Failed to upsert item with hash", "external_id", externalItem.ID, "api_source", req.APISource, "error", err)
			response.FailedCount++
			response.Errors = append(response.Errors, err.Error())
			continue
		}

		item := entity.NewItem()
		item.FromAPIResponse(req.APISource, externalItem)
		response.SuccessCount++
		response.Items = append(response.Items, item)
	}

	if err := uc.cache.Invalidate(ctx, "items:all"); err != nil {
		uc.logger.Warn("Failed to invalidate cache", "error", err, "cache_key", "items:all")
		response.Errors = append(response.Errors, "failed to invalidate cache: "+err.Error())
	}

	return response, nil
}
