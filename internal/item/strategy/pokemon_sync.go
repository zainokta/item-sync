package strategy

import (
	"context"
	"net/url"
	"strconv"

	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/logger"
)

type PokemonSyncStrategy struct {
	logger logger.Logger
}

func NewPokemonSyncStrategy(logger logger.Logger) *PokemonSyncStrategy {
	return &PokemonSyncStrategy{
		logger: logger,
	}
}

func (p *PokemonSyncStrategy) FetchAllItems(ctx context.Context, apiClient ExternalAPIClient, request SyncItemsRequest) ([]entity.ExternalItem, error) {
	var allItems []entity.ExternalItem
	
	offset := 0
	limit := 20
	
	if request.Params != nil {
		if paramOffset, ok := request.Params["offset"].(int); ok {
			offset = paramOffset
		}
		if paramLimit, ok := request.Params["limit"].(int); ok {
			limit = paramLimit
		}
	}

	for {
		params := map[string]interface{}{
			"offset": offset,
			"limit":  limit,
		}
		
		items, err := apiClient.Fetch(ctx, request.APISource, request.Operation, params)
		if err != nil {
			return allItems, err
		}
		
		if len(items) == 0 {
			break
		}
		
		allItems = append(allItems, items...)
		
		nextURL, hasNext := p.extractNextURL(items)
		if !hasNext {
			break
		}
		
		newOffset, newLimit, err := p.parseNextURL(nextURL)
		if err != nil {
			if p.logger != nil {
				p.logger.Warn("Failed to parse next URL, falling back to increment", "url", nextURL, "error", err)
			}
			offset += limit
		} else {
			offset = newOffset
			limit = newLimit
		}
		
		if len(items) < limit {
			break
		}
	}
	
	return allItems, nil
}

func (p *PokemonSyncStrategy) extractNextURL(items []entity.ExternalItem) (string, bool) {
	if len(items) == 0 {
		return "", false
	}
	
	// The Pokemon API client should store the full response in the first item's ExtendInfo
	// Look for pagination info from the response metadata
	for _, item := range items {
		if item.ExtendInfo == nil {
			continue
		}
		
		// Check if this item has pagination info stored by the API client
		if responseData, ok := item.ExtendInfo["response_metadata"]; ok {
			if metadata, ok := responseData.(map[string]interface{}); ok {
				if nextURL, exists := metadata["next"]; exists && nextURL != nil {
					if nextStr, ok := nextURL.(string); ok {
						return nextStr, true
					}
				}
			}
		}
		
		// Fallback: check raw_data
		if rawData, ok := item.ExtendInfo["raw_data"]; ok {
			if pokemonItem, ok := rawData.(map[string]interface{}); ok {
				if nextURL, exists := pokemonItem["next"]; exists && nextURL != nil {
					if nextStr, ok := nextURL.(string); ok {
						return nextStr, true
					}
				}
			}
		}
	}
	
	return "", false
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