package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/logger"
	"github.com/zainokta/item-sync/pkg/retry"
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
	*BaseClient
	apiKey string
}

func NewOpenWeatherClient(config config.APIConfig, retryConfig config.RetryConfig, logger logger.Logger) *OpenWeatherClient {
	return &OpenWeatherClient{
		BaseClient: getBaseClient(config, retryConfig, logger),
		apiKey:     config.OpenWeatherAPIKey,
	}
}

func (c *OpenWeatherClient) Fetch(ctx context.Context, apiName string, operation string, params map[string]interface{}) ([]entity.ExternalItem, error) {
	// Check if API key is configured
	if c.apiKey == "" {
		return nil, errors.ExternalAPIFailed(fmt.Errorf("OpenWeather API key not configured"))
	}

	// Hardcoded OpenWeather API configuration
	baseURL := "https://api.openweathermap.org/data/2.5"

	var endpoint string
	switch operation {
	case "weather":
		endpoint = "/weather"
	default:
		return nil, errors.ExternalAPIFailed(fmt.Errorf("unsupported operation '%s' for OpenWeather API", operation))
	}

	url := fmt.Sprintf("%s%s", baseURL, endpoint)

	// Handle parameters
	if len(params) > 0 {
		if city, ok := params["city"].(string); ok {
			url = fmt.Sprintf("%s?q=%s&appid=%s&units=metric", url, city, c.apiKey)
		}
	}

	var response WeatherResponse
	err := c.doRequest(ctx, http.MethodGet, url, &response)
	if err != nil {
		return nil, errors.ExternalAPIFailed(err)
	}

	return c.transformWeatherResponse(response), nil
}

func (c *OpenWeatherClient) FetchByID(ctx context.Context, apiName string, id int) (entity.ExternalItem, error) {
	// OpenWeather API doesn't support fetch by ID in the same way
	// This method is kept for interface compatibility but returns an error
	return entity.ExternalItem{}, errors.ExternalAPIFailed(fmt.Errorf("FetchByID not supported for OpenWeather API"))
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

func (c *OpenWeatherClient) FetchPaginated(ctx context.Context, apiName string, operation string, params map[string]interface{}) (*PaginatedResponse, error) {
	// Check if API key is configured
	if c.apiKey == "" {
		return nil, errors.ExternalAPIFailed(fmt.Errorf("OpenWeather API key not configured"))
	}

	// OpenWeather API doesn't support pagination like Pokemon API
	// It typically returns single weather data per city
	// This method is kept for interface compatibility

	items, err := c.Fetch(ctx, apiName, operation, params)
	if err != nil {
		return nil, err
	}

	// OpenWeather returns single item, so no pagination metadata
	pagination := NewPaginationMetadata(len(items), "", "")

	return NewPaginatedResponse(items, pagination), nil
}

func (c *OpenWeatherClient) doRequest(ctx context.Context, method, url string, result interface{}) error {
	breaker := c.breakerManager.GetBreaker("openweather-api")

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
