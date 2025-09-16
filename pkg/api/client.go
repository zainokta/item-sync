package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/circuit"
	"github.com/zainokta/item-sync/pkg/logger"
	"github.com/zainokta/item-sync/pkg/retry"
)

type ExternalAPIClient interface {
	Fetch(ctx context.Context, apiName string, operation string, params map[string]interface{}) ([]entity.ExternalItem, error)
	FetchByID(ctx context.Context, apiName string, id int) (entity.ExternalItem, error)
	FetchPaginated(ctx context.Context, apiName string, operation string, params map[string]interface{}) (*PaginatedResponse, error)
}

type BaseClient struct {
	client         *http.Client
	retrier        *retry.Retrier
	breakerManager *circuit.BreakerManager
	logger         logger.Logger
}

func NewAPIClient(apiType string, config config.APIConfig, retryConfig config.RetryConfig, logger logger.Logger) (ExternalAPIClient, error) {
	retrier := retry.New(retryConfig, logger)
	breakerManager := circuit.NewBreakerManager(retryConfig, logger)
	
	switch apiType {
	case "pokemon":
		return NewPokemonClient(config, retrier, breakerManager, logger), nil
	case "openweather":
		return NewOpenWeatherClient(config, retrier, breakerManager, logger), nil
	default:
		return nil, fmt.Errorf("unsupported API type: %s", apiType)
	}
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