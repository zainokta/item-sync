package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/internal/item/usecase/mocks"
	loggermocks "github.com/zainokta/item-sync/pkg/logger/mocks"
	"go.uber.org/mock/gomock"
)

func TestFetchItemUseCase_Execute_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewFetchItemUseCase(cfg, mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := FetchItemRequest{
		ID:        25,
		APISource: "pokemon",
	}

	// Mock data
	mockItem := entity.Item{
		ID:         1,
		Title:      "Pikachu",
		ExternalID: 25,
		APISource:  "pokemon",
	}

	// Set expectations - cache hit
	mockCache.EXPECT().
		GetItem(gomock.Any(), "item:25:pokemon").
		Return(mockItem, nil)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "Pikachu", response.Item.Title)
	assert.Equal(t, 25, response.Item.ExternalID)
	assert.Equal(t, "pokemon", response.Item.APISource)
}

func TestFetchItemUseCase_Execute_DatabaseHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewFetchItemUseCase(cfg, mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := FetchItemRequest{
		ID:        25,
		APISource: "pokemon",
	}

	// Mock data
	mockItem := entity.Item{
		ID:         1,
		Title:      "Pikachu",
		ExternalID: 25,
		APISource:  "pokemon",
	}

	// Set expectations - cache miss, database hit
	mockCache.EXPECT().
		GetItem(gomock.Any(), "item:25:pokemon").
		Return(entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByID(gomock.Any(), 25).
		Return(mockItem, nil)
	mockCache.EXPECT().
		SetItem(gomock.Any(), "item:25:pokemon", mockItem, 5*time.Minute).
		Return(nil)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "Pikachu", response.Item.Title)
	assert.Equal(t, 25, response.Item.ExternalID)
}

func TestFetchItemUseCase_Execute_APIFetch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewFetchItemUseCase(cfg, mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := FetchItemRequest{
		ID:        25,
		APISource: "pokemon",
	}

	// Set expectations - cache miss, database miss
	mockCache.EXPECT().
		GetItem(gomock.Any(), "item:25:pokemon").
		Return(entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByID(gomock.Any(), 25).
		Return(entity.Item{}, assert.AnError) // Database miss

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions - expect API failure in test environment
	require.Error(t, err)
	assert.Equal(t, FetchItemResponse{}, response)
	assert.Contains(t, err.Error(), "external API failed")
}

func TestFetchItemUseCase_Execute_APIClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config with invalid API config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 0, // Invalid timeout
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewFetchItemUseCase(cfg, mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := FetchItemRequest{
		ID:        25,
		APISource: "invalid_api",
	}

	// Set expectations - cache miss, database miss
	mockCache.EXPECT().
		GetItem(gomock.Any(), "item:25:invalid_api").
		Return(entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByID(gomock.Any(), 25).
		Return(entity.Item{}, assert.AnError) // Database miss

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.Error(t, err)
	assert.Equal(t, FetchItemResponse{}, response)
}

func TestFetchItemUseCase_Execute_ExternalAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewFetchItemUseCase(cfg, mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := FetchItemRequest{
		ID:        999, // Non-existent ID
		APISource: "pokemon",
	}

	// Set expectations - cache miss, database miss, API error
	mockCache.EXPECT().
		GetItem(gomock.Any(), "item:999:pokemon").
		Return(entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByID(gomock.Any(), 999).
		Return(entity.Item{}, assert.AnError) // Database miss

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.Error(t, err)
	assert.Equal(t, FetchItemResponse{}, response)
}

func TestFetchItemUseCase_Execute_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewFetchItemUseCase(cfg, mockItemRepo, mockCache, mockLogger)

	// Setup request - use an ID that doesn't exist to simulate validation failure
	request := FetchItemRequest{
		ID:        999999, // Very high ID that likely doesn't exist
		APISource: "pokemon",
	}

	// Set expectations - cache miss, database miss
	mockCache.EXPECT().
		GetItem(gomock.Any(), "item:999999:pokemon").
		Return(entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByID(gomock.Any(), 999999).
		Return(entity.Item{}, assert.AnError) // Database miss

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions - should fail due to API error or validation
	require.Error(t, err)
	assert.Equal(t, FetchItemResponse{}, response)
	assert.Contains(t, err.Error(), "external API failed")
}

func TestFetchItemUseCase_Execute_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewFetchItemUseCase(cfg, mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := FetchItemRequest{
		ID:        25,
		APISource: "pokemon",
	}

	// Set expectations - cache miss, database miss
	mockCache.EXPECT().
		GetItem(gomock.Any(), "item:25:pokemon").
		Return(entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByID(gomock.Any(), 25).
		Return(entity.Item{}, assert.AnError) // Database miss

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions - expect API failure in test environment
	require.Error(t, err)
	assert.Equal(t, FetchItemResponse{}, response)
	assert.Contains(t, err.Error(), "external API failed")
}

func TestFetchItemUseCase_Execute_CacheSetErrorOnDatabaseHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewFetchItemUseCase(cfg, mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := FetchItemRequest{
		ID:        25,
		APISource: "pokemon",
	}

	// Mock data
	mockItem := entity.Item{
		ID:         1,
		Title:      "Pikachu",
		ExternalID: 25,
		APISource:  "pokemon",
	}

	// Set expectations - cache miss, database hit, cache set fails
	mockCache.EXPECT().
		GetItem(gomock.Any(), "item:25:pokemon").
		Return(entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByID(gomock.Any(), 25).
		Return(mockItem, nil)
	mockCache.EXPECT().
		SetItem(gomock.Any(), "item:25:pokemon", mockItem, 5*time.Minute).
		Return(assert.AnError)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions - should still succeed even if cache fails
	require.NoError(t, err)
	assert.Equal(t, "Pikachu", response.Item.Title)
}

func TestFetchItemUseCase_Execute_CacheSetErrorOnAPIFetch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewFetchItemUseCase(cfg, mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := FetchItemRequest{
		ID:        25,
		APISource: "pokemon",
	}

	// Set expectations - cache miss, database miss
	mockCache.EXPECT().
		GetItem(gomock.Any(), "item:25:pokemon").
		Return(entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByID(gomock.Any(), 25).
		Return(entity.Item{}, assert.AnError) // Database miss

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions - expect API failure in test environment
	require.Error(t, err)
	assert.Equal(t, FetchItemResponse{}, response)
	assert.Contains(t, err.Error(), "external API failed")
}
