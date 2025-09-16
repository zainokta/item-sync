package usecase

import (
	"context"
	"errors"
	"time"

	pkgErrors "github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/logger"
)

type SyncItemsUseCase struct {
	itemRepo  ItemRepository
	apiClient ExternalAPIClient
	cache     ItemCache
	logger    logger.Logger
}

func NewSyncItemsUseCase(itemRepo ItemRepository, apiClient ExternalAPIClient, cache ItemCache, logger logger.Logger) *SyncItemsUseCase {
	return &SyncItemsUseCase{
		itemRepo:  itemRepo,
		apiClient: apiClient,
		cache:     cache,
		logger:    logger,
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
	externalItems, err := uc.apiClient.Fetch(ctx, req.APISource, req.Operation, req.Params)
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
		item := entity.NewItem()
		item.FromAPIResponse(externalItem)

		if err := item.Validate(); err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors, err.Error())
			continue
		}

		existingItem, err := uc.itemRepo.FindByExternalID(ctx, externalItem.ID)
		if err != nil {
			if errors.Is(err, pkgErrors.ItemNotFound()) {
				if err := uc.itemRepo.Save(ctx, item); err != nil {
					response.FailedCount++
					response.Errors = append(response.Errors, err.Error())
					continue
				}
			} else {
				response.FailedCount++
				response.Errors = append(response.Errors, err.Error())
				continue
			}
		} else {
			if !req.ForceSync {
				continue
			}

			item.ID = existingItem.ID
			item.CreatedAt = existingItem.CreatedAt
			item.UpdatedAt = time.Now()
			if err := uc.itemRepo.Update(ctx, item); err != nil {
				response.FailedCount++
				response.Errors = append(response.Errors, err.Error())
				continue
			}
		}

		response.SuccessCount++
		response.Items = append(response.Items, item)
	}

	if err := uc.cache.Invalidate(ctx, "items:all"); err != nil {
		uc.logger.Warn("Failed to invalidate cache", "error", err, "cache_key", "items:all")
		response.Errors = append(response.Errors, "failed to invalidate cache: "+err.Error())
	}

	return response, nil
}
