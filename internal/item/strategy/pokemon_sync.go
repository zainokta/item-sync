package strategy

import (
	"context"
	"net/url"
	"strconv"

	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/logger"
)

type PokemonSyncStrategy struct {
	logger    logger.Logger
	apiClient ExternalAPIClient
}

func NewPokemonSyncStrategy(logger logger.Logger, apiClient ExternalAPIClient) *PokemonSyncStrategy {
	return &PokemonSyncStrategy{
		logger:    logger,
		apiClient: apiClient,
	}
}

func (p *PokemonSyncStrategy) FetchAllItems(ctx context.Context, request SyncItemsRequest) ([]entity.ExternalItem, error) {
	var allItems []entity.ExternalItem

	offset := 0
	limit := 20

	if request.Params != nil {
		if paramOffset, ok := request.Params["offset"].(float64); ok {
			offset = int(paramOffset)
		}
		if paramLimit, ok := request.Params["limit"].(float64); ok {
			limit = int(paramLimit)
		}
	}

	for {
		params := map[string]interface{}{
			"offset": offset,
			"limit":  limit,
		}

		response, err := p.apiClient.FetchPaginated(ctx, request.APISource, request.Operation, params)
		if err != nil {
			return allItems, err
		}

		if len(response.Items) == 0 {
			break
		}

		allItems = append(allItems, response.Items...)

		// Use structured pagination metadata
		if response.Pagination == nil || !response.Pagination.HasNext {
			break
		}

		// Parse next URL to get new offset and limit
		if response.Pagination.Next != "" {
			newOffset, newLimit, err := p.parseNextURL(response.Pagination.Next)
			if err != nil {
				if p.logger != nil {
					p.logger.Warn("Failed to parse next URL, falling back to increment", "url", response.Pagination.Next, "error", err)
				}
				offset += limit
			} else {
				offset = newOffset
				limit = newLimit
			}
		} else {
			offset += limit
		}

		// Safety check to prevent infinite loops
		if len(response.Items) < limit {
			break
		}
	}

	return allItems, nil
}

func (p *PokemonSyncStrategy) parseNextURL(nextURL string) (offset, limit int, err error) {
	parsedURL, err := url.Parse(nextURL)
	if err != nil {
		return 0, 20, err
	}

	query := parsedURL.Query()

	offsetStr := query.Get("offset")
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			offset = 0
		}
	}

	limitStr := query.Get("limit")
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			limit = 20
		}
	} else {
		limit = 20
	}

	return offset, limit, nil
}

func (p *PokemonSyncStrategy) Fetch(ctx context.Context, request SyncItemsRequest) ([]entity.ExternalItem, error) {
	offset := 0
	limit := 20

	if request.Params != nil {
		if paramOffset, ok := request.Params["offset"].(float64); ok {
			offset = int(paramOffset)
		}
		if paramLimit, ok := request.Params["limit"].(float64); ok {
			limit = int(paramLimit)
		}
	}

	// Make single API call with specified parameters
	params := map[string]interface{}{
		"offset": offset,
		"limit":  limit,
	}

	response, err := p.apiClient.FetchPaginated(ctx, request.APISource, request.Operation, params)
	if err != nil {
		return nil, err
	}

	return response.Items, nil
}
