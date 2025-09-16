package api

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/circuit"
	"github.com/zainokta/item-sync/pkg/logger"
	"github.com/zainokta/item-sync/pkg/retry"
)

type PokemonResponse struct {
	Count    int           `json:"count"`
	Next     string        `json:"next"`
	Previous string        `json:"previous"`
	Results  []PokemonItem `json:"results"`
}

type PokemonItem struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokemonClient struct {
	*BaseClient
}

func NewPokemonClient(config config.APIConfig, retrier *retry.Retrier, breakerManager *circuit.BreakerManager, logger logger.Logger) *PokemonClient {
	return &PokemonClient{
		BaseClient: &BaseClient{
			client: &http.Client{
				Timeout: config.Timeout,
				Transport: &http.Transport{
					MaxIdleConns:        config.MaxIdleConns,
					IdleConnTimeout:     config.IdleConnTimeout,
					DisableCompression:  config.DisableCompression,
					MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
				},
			},
			retrier:        retrier,
			breakerManager: breakerManager,
			logger:         logger,
		},
	}
}

func (c *PokemonClient) Fetch(ctx context.Context, apiName string, operation string, params map[string]interface{}) ([]entity.ExternalItem, error) {
	// Hardcoded Pokemon API configuration
	baseURL := "https://pokeapi.co/api/v2"
	
	var endpoint string
	switch operation {
	case "list":
		endpoint = "/pokemon"
	case "get":
		endpoint = "/pokemon"
	default:
		return nil, errors.ExternalAPIFailed(fmt.Errorf("unsupported operation '%s' for Pokemon API", operation))
	}

	url := fmt.Sprintf("%s%s", baseURL, endpoint)

	// Handle parameters
	if len(params) > 0 {
		if id, ok := params["id"].(int); ok {
			url = fmt.Sprintf("%s/%d", url, id)
		} else {
			// Handle list parameters (limit, offset)
			queryParams := ""
			if limit, ok := params["limit"].(int); ok {
				queryParams += fmt.Sprintf("limit=%d", limit)
			}
			if offset, ok := params["offset"].(int); ok {
				if queryParams != "" {
					queryParams += "&"
				}
				queryParams += fmt.Sprintf("offset=%d", offset)
			}
			if queryParams != "" {
				url = fmt.Sprintf("%s?%s", url, queryParams)
			}
		}
	}

	var response PokemonResponse
	err := c.doRequest(ctx, http.MethodGet, url, &response)
	if err != nil {
		return nil, errors.ExternalAPIFailed(err)
	}

	return c.transformPokemonResponse(response), nil
}

func (c *PokemonClient) FetchByID(ctx context.Context, apiName string, id int) (entity.ExternalItem, error) {
	params := map[string]interface{}{"id": id}
	items, err := c.Fetch(ctx, apiName, "get", params)
	if err != nil {
		return entity.ExternalItem{}, err
	}

	if len(items) == 0 {
		return entity.ExternalItem{}, errors.ExternalAPIFailed(fmt.Errorf("pokemon not found"))
	}

	return items[0], nil
}

func (c *PokemonClient) transformPokemonResponse(response PokemonResponse) []entity.ExternalItem {
	var items []entity.ExternalItem

	for _, pokemonItem := range response.Results {
		externalItem := entity.ExternalItem{
			ID:         c.extractPokemonID(pokemonItem.URL),
			Title:      pokemonItem.Name,
			ExtendInfo: make(map[string]interface{}),
		}

		externalItem.ExtendInfo["raw_data"] = pokemonItem

		items = append(items, externalItem)
	}

	return items
}

func (c *PokemonClient) FetchPaginated(ctx context.Context, apiName string, operation string, params map[string]interface{}) (*PaginatedResponse, error) {
	// Hardcoded Pokemon API configuration
	baseURL := "https://pokeapi.co/api/v2"
	
	var endpoint string
	switch operation {
	case "list":
		endpoint = "/pokemon"
	case "get":
		endpoint = "/pokemon"
	default:
		return nil, errors.ExternalAPIFailed(fmt.Errorf("unsupported operation '%s' for Pokemon API", operation))
	}

	url := fmt.Sprintf("%s%s", baseURL, endpoint)

	// Handle parameters
	if len(params) > 0 {
		if id, ok := params["id"].(int); ok {
			url = fmt.Sprintf("%s/%d", url, id)
		} else {
			// Handle list parameters (limit, offset)
			queryParams := ""
			if limit, ok := params["limit"].(int); ok {
				queryParams += fmt.Sprintf("limit=%d", limit)
			}
			if offset, ok := params["offset"].(int); ok {
				if queryParams != "" {
					queryParams += "&"
				}
				queryParams += fmt.Sprintf("offset=%d", offset)
			}
			if queryParams != "" {
				url = fmt.Sprintf("%s?%s", url, queryParams)
			}
		}
	}

	var response PokemonResponse
	err := c.doRequest(ctx, http.MethodGet, url, &response)
	if err != nil {
		return nil, errors.ExternalAPIFailed(err)
	}

	items := c.transformPokemonResponse(response)
	pagination := NewPaginationMetadata(response.Count, response.Next, response.Previous)
	
	return NewPaginatedResponse(items, pagination), nil
}

func (c *PokemonClient) extractPokemonID(url string) int {
	re := regexp.MustCompile(`/pokemon/(\d+)/?`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		if id, err := strconv.Atoi(matches[1]); err == nil {
			return id
		}
	}
	return 0
}

func (c *PokemonClient) doRequest(ctx context.Context, method, url string, result interface{}) error {
	breaker := c.breakerManager.GetBreaker("pokemon-api")
	
	return breaker.Execute(func() error {
		return c.retrier.Execute(ctx, func() error {
			req, err := http.NewRequestWithContext(ctx, method, url, nil)
			if err != nil {
				return retry.NewNonRetryableError(err)
			}

			req.Header.Set("Accept", "application/json")

			err = c.BaseClient.doRequest(req, result)
			if err != nil {
				if shouldNotRetry(err) {
					return retry.NewNonRetryableError(err)
				}
				return retry.NewRetryableError(err)
			}
			
			return nil
		})
	})
}
