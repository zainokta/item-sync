package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/logger"
)

type ListItemsUseCase struct {
	itemRepo ItemRepository
	cache    ItemCache
	logger   logger.Logger
}

func NewListItemsUseCase(itemRepo ItemRepository, cache ItemCache, logger logger.Logger) *ListItemsUseCase {
	return &ListItemsUseCase{
		itemRepo: itemRepo,
		cache:    cache,
		logger:   logger,
	}
}

type ListItemsRequest struct {
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	ItemType  string `json:"item_type"`
	Status    string `json:"status"`
	APISource string `json:"api_source"`
}

type ListItemsResponse struct {
	Items      []entity.Item `json:"items"`
	TotalCount int           `json:"total_count"`
}

func (uc *ListItemsUseCase) Execute(ctx context.Context, req ListItemsRequest) (ListItemsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	cacheKey := fmt.Sprintf("items:%s:%s:%d:%d", req.APISource, req.Status, req.Limit, req.Offset)

	if cachedItems, err := uc.cache.GetItems(ctx, cacheKey); err == nil {
		return ListItemsResponse{
			Items: cachedItems,
		}, nil
	}

	var items []entity.Item
	var err error

	switch {
	case req.APISource != "":
		items, err = uc.itemRepo.FindByAPISource(ctx, req.APISource, req.Limit, req.Offset)
	case req.Status != "":
		items, err = uc.itemRepo.FindByStatus(ctx, req.Status, req.Limit, req.Offset)
	case req.ItemType != "":
		items, err = uc.itemRepo.FindByType(ctx, req.ItemType, req.Limit, req.Offset)
	default:
		items, err = uc.itemRepo.FindAll(ctx, req.Limit, req.Offset)
	}

	if err != nil {
		return ListItemsResponse{}, errors.DatabaseError(err)
	}

	if len(items) != 0 {
		if cacheErr := uc.cache.SetItems(ctx, cacheKey, items, 10*time.Minute); cacheErr != nil {
			uc.logger.Warn("Failed to cache items", "error", cacheErr, "cache_key", cacheKey)
		}
	}

	return ListItemsResponse{
		Items:      items,
		TotalCount: len(items),
	}, nil
}
