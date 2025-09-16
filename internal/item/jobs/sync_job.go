package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/item/usecase"
	"github.com/zainokta/item-sync/pkg/logger"
)

type SyncJob struct {
	name           string
	itemRepository usecase.ItemRepository
	jobRepository  JobRepository
	cache          usecase.ItemCache
	apiClient      usecase.ExternalAPIClient
	apiType        string
	logger         logger.Logger
	config         config.Config
}

func NewSyncJob(
	name string,
	itemRepository usecase.ItemRepository,
	jobRepository JobRepository,
	cache usecase.ItemCache,
	apiClient usecase.ExternalAPIClient,
	apiType string,
	logger logger.Logger,
	config config.Config,
) *SyncJob {
	return &SyncJob{
		name:           name,
		itemRepository: itemRepository,
		jobRepository:  jobRepository,
		cache:          cache,
		apiClient:      apiClient,
		apiType:        apiType,
		logger:         logger,
		config:         config,
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
	limit := 20
	offset := 0

	for {
		params := map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		}

		items, err := j.apiClient.Fetch(ctx, "pokemon", "list", params)
		if err != nil {
			j.logger.Error("Failed to fetch Pokemon data", "error", err, "offset", offset)
			lastErr = err
			break
		}

		if len(items) == 0 {
			j.logger.Info("No more Pokemon data to sync")
			break
		}

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
		}

		if len(items) < limit {
			j.logger.Info("Reached end of Pokemon data")
			break
		}

		offset += limit

		select {
		case <-ctx.Done():
			lastErr = ctx.Err()
			return
		default:
		}
	}

	return
}

func (j *SyncJob) syncWeatherData(ctx context.Context) (processed, succeeded, failed int, lastErr error) {
	cities := []string{"Jakarta", "Bandung", "Surabaya"}

	for _, city := range cities {
		params := map[string]interface{}{
			"city": city,
		}

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
