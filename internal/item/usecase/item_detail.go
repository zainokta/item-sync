package usecase

import (
	"context"
	"fmt"
	"time"

	pkgErrors "github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/logger"
)

type FetchItemUseCase struct {
	itemRepo  ItemRepository
	apiClient ExternalAPIClient
	cache     ItemCache
	logger    logger.Logger
}

type FetchItemRequest struct {
	ID        int    `json:"id"`
	APISource string `json:"api_source"`
}

type FetchItemResponse struct {
	Item      entity.Item `json:"item"`
	FromCache bool        `json:"from_cache"`
}

func NewFetchItemUseCase(itemRepo ItemRepository, apiClient ExternalAPIClient, cache ItemCache, logger logger.Logger) *FetchItemUseCase {
	return &FetchItemUseCase{
		itemRepo:  itemRepo,
		apiClient: apiClient,
		cache:     cache,
		logger:    logger,
	}
}

func (uc *FetchItemUseCase) Execute(ctx context.Context, req FetchItemRequest) (FetchItemResponse, error) {
	cacheKey := fmt.Sprintf("item:%d:%s", req.ID, req.APISource)
	if cachedItem, err := uc.cache.GetItem(ctx, cacheKey); err == nil {
		return FetchItemResponse{
			Item:      cachedItem,
			FromCache: true,
		}, nil
	}

	if item, err := uc.itemRepo.FindByID(ctx, req.ID); err == nil {
		if cacheErr := uc.cache.SetItem(ctx, cacheKey, item, 5*time.Minute); cacheErr != nil {
			uc.logger.Warn("Failed to cache item", "error", cacheErr, "cache_key", cacheKey)
		}
		return FetchItemResponse{
			Item:      item,
			FromCache: false,
		}, nil
	}

	externalItem, err := uc.apiClient.FetchByID(ctx, req.APISource, req.ID)
	if err != nil {
		return FetchItemResponse{}, pkgErrors.ExternalAPIFailed(err)
	}

	item := entity.NewItem()
	item.FromAPIResponse(externalItem)

	if err := item.Validate(); err != nil {
		return FetchItemResponse{}, err
	}

	if err := uc.itemRepo.Save(ctx, item); err != nil {
		return FetchItemResponse{}, err
	}

	if cacheErr := uc.cache.SetItem(ctx, cacheKey, item, 5*time.Minute); cacheErr != nil {
		uc.logger.Warn("Failed to cache item", "error", cacheErr, "cache_key", cacheKey)
	}

	return FetchItemResponse{
		Item:      item,
		FromCache: false,
	}, nil
}
