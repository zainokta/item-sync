package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/item/entity"
)

type ExternalAPIClient interface {
	Fetch(ctx context.Context, apiName string, operation string, params map[string]interface{}) ([]entity.ExternalItem, error)
	FetchByID(ctx context.Context, apiName string, id int) (entity.ExternalItem, error)
}

type BaseClient struct {
	client *http.Client
}

func NewAPIClient(apiType string, config config.APIConfig) (ExternalAPIClient, error) {
	switch apiType {
	case "pokemon":
		return NewPokemonClient(config), nil
	case "openweather":
		return NewOpenWeatherClient(config), nil
	default:
		return nil, fmt.Errorf("unsupported API type: %s", apiType)
	}
}

func (bc *BaseClient) doRequestWithAuth(ctx context.Context, apiConfig config.ExternalAPIConfig, method, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	switch apiConfig.AuthType {
	case "bearer":
		if apiConfig.APIKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiConfig.APIKey))
		}
	case "api_key":
		if apiConfig.APIKey != "" {
			req.Header.Set("X-API-Key", apiConfig.APIKey)
		}
	}

	for key, value := range apiConfig.Headers {
		req.Header.Set(key, value)
	}

	var lastErr error
	for i := 0; i < apiConfig.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := bc.doRequest(req, result)
		if err == nil {
			return nil
		}

		lastErr = err

		if shouldNotRetry(err) {
			break
		}

		if i < apiConfig.MaxRetries-1 {
			select {
			case <-time.After(apiConfig.RetryDelay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return lastErr
}

func (bc *BaseClient) doRequest(req *http.Request, result interface{}) error {
	resp, err := bc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func shouldNotRetry(err error) bool {
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}

	errStr := err.Error()

	if len(errStr) >= 8 && errStr[:4] == "HTTP" {
		if len(errStr) >= 8 {
			statusStr := errStr[5:8]
			if statusStr[0] == '4' {
				return statusStr != "408" && statusStr != "429"
			}
		}
	}

	return false
}