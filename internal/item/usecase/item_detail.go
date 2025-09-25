package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/zainokta/item-sync/config"
	pkgErrors "github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/api"
	"github.com/zainokta/item-sync/pkg/logger"
)

type FetchItemUseCase struct {
	cfg      *config.Config
	itemRepo ItemRepository
	cache    ItemCache
	logger   logger.Logger
}

type FetchItemRequest struct {
	ID        int    `json:"id"`
	APISource string `json:"api_source"`
}

type FetchItemResponse struct {
	Item entity.Item `json:"item"`
}

func NewFetchItemUseCase(cfg *config.Config, itemRepo ItemRepository, cache ItemCache, logger logger.Logger) *FetchItemUseCase {
	return &FetchItemUseCase{
		cfg:      cfg,
		itemRepo: itemRepo,
		cache:    cache,
		logger:   logger,
	}
}

func (uc *FetchItemUseCase) Execute(ctx context.Context, req FetchItemRequest) (FetchItemResponse, error) {
	cacheKey := fmt.Sprintf("item:%d:%s", req.ID, req.APISource)
	if cachedItem, err := uc.cache.GetItem(ctx, cacheKey); err == nil {
		return FetchItemResponse{
			Item: cachedItem,
		}, nil
	}

	if item, err := uc.itemRepo.FindByID(ctx, req.ID); err == nil {
		if cacheErr := uc.cache.SetItem(ctx, cacheKey, item, 5*time.Minute); cacheErr != nil {
			uc.logger.Warn("Failed to cache item", "error", cacheErr, "cache_key", cacheKey)
		}
		return FetchItemResponse{
			Item: item,
		}, nil
	}

	apiClient, err := api.NewAPIClient(req.APISource, uc.cfg.API, uc.cfg.Retry, uc.logger)
	if err != nil {
		return FetchItemResponse{}, err
	}

	externalItem, err := apiClient.FetchByID(ctx, req.APISource, req.ID)
	if err != nil {
		return FetchItemResponse{}, pkgErrors.ExternalAPIFailed(err)
	}

	item := entity.NewItem()
	item.FromAPIResponse(req.APISource, externalItem)

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
		Item: item,
	}, nil
}
