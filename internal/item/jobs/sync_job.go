package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/internal/item/strategy"
	"github.com/zainokta/item-sync/pkg/logger"
)

type SyncJob struct {
	name           string
	itemRepository ItemSaver
	jobRepository  JobRepository
	apiClient      ExternalAPIClient
	apiType        string
	logger         logger.Logger
	config         config.Config
	params         map[string]interface{}
}

func NewSyncJob(
	name string,
	itemRepository ItemSaver,
	jobRepository JobRepository,
	apiClient ExternalAPIClient,
	apiType string,
	logger logger.Logger,
	config config.Config,
	params map[string]interface{},
) *SyncJob {
	// Default to empty map if params is nil
	if params == nil {
		params = make(map[string]interface{})
	}

	return &SyncJob{
		name:           name,
		itemRepository: itemRepository,
		jobRepository:  jobRepository,
		apiClient:      apiClient,
		apiType:        apiType,
		logger:         logger,
		config:         config,
		params:         params,
	}
}

func (j *SyncJob) Name() string {
	return j.name
}

func (j *SyncJob) Execute(ctx context.Context) error {
	if j.apiClient == nil {
		return fmt.Errorf("API client not configured for %s", j.apiType)
	}

	j.logger.Info("Starting background sync job", "api_type", j.apiType)

	jobID, err := j.jobRepository.CreateSyncJobRecord(ctx, j.name, j.apiType)
	if err != nil {
		j.logger.Error("Failed to create sync job record", "error", err)
		return err
	}

	startTime := time.Now()
	itemsProcessed := 0
	itemsSucceeded := 0
	itemsFailed := 0
	var lastError error

	defer func() {
		executionTime := time.Since(startTime)
		status := "completed"
		if lastError != nil {
			status = "failed"
		}

		err := j.jobRepository.UpdateSyncJobRecord(ctx, jobID, status, itemsProcessed, itemsSucceeded, itemsFailed, lastError, executionTime)
		if err != nil {
			j.logger.Error("Failed to update sync job record", "error", err)
		}
	}()

	switch j.apiType {
	case "pokemon":
		itemsProcessed, itemsSucceeded, itemsFailed, lastError = j.syncPokemonData(ctx)
	case "openweather":
		itemsProcessed, itemsSucceeded, itemsFailed, lastError = j.syncWeatherData(ctx)
	default:
		lastError = fmt.Errorf("unsupported API type: %s", j.apiType)
		return lastError
	}

	if lastError != nil {
		j.logger.Error("Sync job completed with errors",
			"processed", itemsProcessed,
			"succeeded", itemsSucceeded,
			"failed", itemsFailed,
			"error", lastError)
		return lastError
	}

	j.logger.Info("Sync job completed successfully",
		"processed", itemsProcessed,
		"succeeded", itemsSucceeded,
		"failed", itemsFailed)

	return nil
}

func (j *SyncJob) syncPokemonData(ctx context.Context) (processed, succeeded, failed int, lastErr error) {
	pokemonStrategy := strategy.NewPokemonSyncStrategy(j.logger, j.apiClient)

	request := strategy.SyncItemsRequest{
		APISource: "pokemon",
		Operation: "list",
		Params:    j.params,
	}

	var items []entity.ExternalItem
	var err error

	// Check if user wants limited fetch or full sync
	if _, hasLimit := j.params["limit"]; hasLimit {
		j.logger.Info("Fetching limited Pokemon data", "params", j.params)
		items, err = pokemonStrategy.Fetch(ctx, request)
	} else {
		j.logger.Info("Fetching all Pokemon data")
		items, err = pokemonStrategy.FetchAllItems(ctx, request)
	}

	if err != nil {
		j.logger.Error("Failed to fetch Pokemon data using strategy", "error", err)
		lastErr = err
		return
	}

	j.logger.Info("Fetched Pokemon data successfully", "total_items", len(items))

	for _, item := range items {
		processed++

		err := j.itemRepository.UpsertWithHash(ctx, "pokemon", item)
		if err != nil {
			j.logger.Error("Failed to store Pokemon item", "id", item.ID, "error", err)
			failed++
			lastErr = err
		} else {
			succeeded++
			j.logger.Debug("Successfully stored Pokemon item", "id", item.ID, "title", item.Title)
		}

		if processed%100 == 0 {
			select {
			case <-ctx.Done():
				lastErr = ctx.Err()
				return
			default:
			}
		}
	}

	j.logger.Info("Pokemon data sync completed", "processed", processed, "succeeded", succeeded, "failed", failed)
	return
}

func (j *SyncJob) syncWeatherData(ctx context.Context) (processed, succeeded, failed int, lastErr error) {
	// Get cities from params or use defaults
	cities := []string{"Jakarta", "Bandung", "Surabaya"}
	if citiesParam, ok := j.params["cities"].(string); ok && len(citiesParam) > 0 {
		cities = strings.Split(citiesParam, ",")
	}

	for _, city := range cities {
		// Merge job params with city-specific params
		params := make(map[string]interface{})
		for k, v := range j.params {
			params[k] = v
		}
		params["city"] = city

		items, err := j.apiClient.Fetch(ctx, "openweather", "weather", params)
		if err != nil {
			j.logger.Error("Failed to fetch weather data", "error", err, "city", city)
			failed++
			lastErr = err
			continue
		}

		for _, item := range items {
			processed++

			err := j.itemRepository.UpsertWithHash(ctx, "openweather", item)
			if err != nil {
				j.logger.Error("Failed to store weather item", "id", item.ID, "error", err)
				failed++
				lastErr = err
			} else {
				succeeded++
				j.logger.Debug("Successfully stored weather item", "id", item.ID, "title", item.Title)
			}
		}

		select {
		case <-ctx.Done():
			lastErr = ctx.Err()
			return
		default:
		}
	}

	return
}
