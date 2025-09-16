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
	client       *http.Client
	externalAPIs map[string]config.ExternalAPIConfig
}

func NewPokemonClient(config config.APIConfig) *PokemonClient {
	return &PokemonClient{
		client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        config.MaxIdleConns,
				IdleConnTimeout:     config.IdleConnTimeout,
				DisableCompression:  config.DisableCompression,
				MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
			},
		},
		externalAPIs: config.ExternalAPIs,
	}
}

func (c *PokemonClient) Fetch(ctx context.Context, apiName string, operation string, params map[string]interface{}) ([]entity.ExternalItem, error) {
	apiConfig, exists := c.externalAPIs[apiName]
	if !exists {
		return nil, errors.ExternalAPIFailed(fmt.Errorf("API '%s' not configured", apiName))
	}

	if !apiConfig.Enable {
		return nil, errors.ExternalAPIFailed(fmt.Errorf("API '%s' is disabled", apiName))
	}

	endpoint, exists := apiConfig.Endpoints[operation]
	if !exists {
		return nil, errors.ExternalAPIFailed(fmt.Errorf("endpoint '%s' not configured for API '%s'", operation, apiName))
	}

	url := fmt.Sprintf("%s%s", apiConfig.BaseURL, endpoint)

	if len(params) > 0 {
		if id, ok := params["id"].(int); ok {
			url = fmt.Sprintf("%s/%d", url, id)
		}
	}

	var response PokemonResponse
	err := c.doRequestWithAuth(ctx, apiConfig, http.MethodGet, url, &response)
	if err != nil {
		return nil, errors.ExternalAPIFailed(err)
	}

	return c.transformPokemonResponse(response, apiConfig), nil
}

func (c *PokemonClient) FetchByID(ctx context.Context, apiName string, id int) (entity.ExternalItem, error) {
	params := map[string]interface{}{"id": id}
	items, err := c.Fetch(ctx, apiName, http.MethodGet, params)
	if err != nil {
		return entity.ExternalItem{}, err
	}

	if len(items) == 0 {
		return entity.ExternalItem{}, errors.ExternalAPIFailed(fmt.Errorf("pokemon not found"))
	}

	return items[0], nil
}

func (c *PokemonClient) transformPokemonResponse(response PokemonResponse, apiConfig config.ExternalAPIConfig) []entity.ExternalItem {
	var items []entity.ExternalItem

	for _, pokemonItem := range response.Results {
		externalItem := entity.ExternalItem{
			ID:         c.extractPokemonID(pokemonItem.URL),
			Title:      pokemonItem.Name,
			ExtendInfo: make(map[string]interface{}),
		}

		externalItem.ExtendInfo["api_source"] = "pokemon"
		externalItem.ExtendInfo["url"] = pokemonItem.URL
		externalItem.ExtendInfo["raw_data"] = pokemonItem

		items = append(items, externalItem)
	}

	return items
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

func (c *PokemonClient) doRequestWithAuth(ctx context.Context, apiConfig config.ExternalAPIConfig, method, url string, result interface{}) error {
	baseClient := &BaseClient{client: c.client}
	return baseClient.doRequestWithAuth(ctx, apiConfig, method, url, result)
}
