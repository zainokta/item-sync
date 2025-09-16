package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
)

type WeatherResponse struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type OpenWeatherClient struct {
	client       *http.Client
	externalAPIs map[string]config.ExternalAPIConfig
}

func NewOpenWeatherClient(config config.APIConfig) *OpenWeatherClient {
	return &OpenWeatherClient{
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

func (c *OpenWeatherClient) Fetch(ctx context.Context, apiName string, operation string, params map[string]interface{}) ([]entity.ExternalItem, error) {
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
		if city, ok := params["city"].(string); ok {
			url = fmt.Sprintf("%s?q=%s&appid=%s", url, city, apiConfig.APIKey)
		}
	}

	var response WeatherResponse
	err := c.doRequestWithAuth(ctx, apiConfig, http.MethodGet, url, &response)
	if err != nil {
		return nil, errors.ExternalAPIFailed(err)
	}

	return c.transformWeatherResponse(response), nil
}

func (c *OpenWeatherClient) FetchByID(ctx context.Context, apiName string, id int) (entity.ExternalItem, error) {
	params := map[string]interface{}{"id": id}
	items, err := c.Fetch(ctx, apiName, "get", params)
	if err != nil {
		return entity.ExternalItem{}, err
	}

	if len(items) == 0 {
		return entity.ExternalItem{}, errors.ExternalAPIFailed(fmt.Errorf("weather data not found"))
	}

	return items[0], nil
}

func (c *OpenWeatherClient) transformWeatherResponse(response WeatherResponse) []entity.ExternalItem {
	var items []entity.ExternalItem

	externalItem := entity.ExternalItem{
		ID:         response.ID,
		Title:      response.Name,
		ExtendInfo: make(map[string]interface{}),
	}

	externalItem.ExtendInfo["api_source"] = "openweather"
	externalItem.ExtendInfo["temperature"] = response.Main.Temp
	externalItem.ExtendInfo["humidity"] = response.Main.Humidity
	if len(response.Weather) > 0 {
		externalItem.ExtendInfo["weather_main"] = response.Weather[0].Main
		externalItem.ExtendInfo["description"] = response.Weather[0].Description
	}
	externalItem.ExtendInfo["raw_data"] = response

	items = append(items, externalItem)
	return items
}

func (c *OpenWeatherClient) doRequestWithAuth(ctx context.Context, apiConfig config.ExternalAPIConfig, method, url string, result interface{}) error {
	baseClient := &BaseClient{client: c.client}
	return baseClient.doRequestWithAuth(ctx, apiConfig, method, url, result)
}