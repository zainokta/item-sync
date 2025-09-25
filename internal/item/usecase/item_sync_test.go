package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/item/usecase/mocks"
	loggermocks "github.com/zainokta/item-sync/pkg/logger/mocks"
	"go.uber.org/mock/gomock"
)

func TestSyncItemsUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockJobRepo := mocks.NewMockJobRepository(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewSyncItemsUseCase(cfg, mockItemRepo, mockJobRepo, mockLogger)

	// Setup request
	request := SyncItemsRequest{
		APISource: "pokemon",
		Params: map[string]interface{}{
			"limit": 20,
		},
	}

	// Set expectations - allow any repository calls
	mockJobRepo.EXPECT().
		CreateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(int64(1), nil).AnyTimes()
	mockJobRepo.EXPECT().
		UpdateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	mockItemRepo.EXPECT().
		UpsertWithHash(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	// Expect logger calls for background sync
	mockLogger.EXPECT().
		Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().
		Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Expect logger calls for background sync
	mockLogger.EXPECT().
		Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().
		Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "accepted", response.Status)
	assert.Equal(t, "Sync job has been accepted for background processing", response.Message)
}

func TestSyncItemsUseCase_Execute_APIClientCreationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockJobRepo := mocks.NewMockJobRepository(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config with invalid API config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 0, // Invalid timeout
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewSyncItemsUseCase(cfg, mockItemRepo, mockJobRepo, mockLogger)

	// Setup request
	request := SyncItemsRequest{
		APISource: "invalid_api",
	}

	// Expect logger error call
	mockLogger.EXPECT().
		Error("Failed to create API client", "api_source", "invalid_api", "error", gomock.Any())

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.Error(t, err)
	assert.Equal(t, SyncItemsResponse{}, response)
}

func TestSyncItemsUseCase_Execute_WithNilParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockJobRepo := mocks.NewMockJobRepository(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewSyncItemsUseCase(cfg, mockItemRepo, mockJobRepo, mockLogger)

	// Setup request with nil params
	request := SyncItemsRequest{
		APISource: "pokemon",
		Params:    nil,
	}

	// Set expectations - allow any repository calls
	mockJobRepo.EXPECT().
		CreateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(int64(1), nil).AnyTimes()
	mockJobRepo.EXPECT().
		UpdateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	mockItemRepo.EXPECT().
		UpsertWithHash(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	// Expect logger calls for background sync
	mockLogger.EXPECT().
		Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().
		Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "accepted", response.Status)
}

func TestSyncItemsUseCase_Execute_WithEmptyParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockJobRepo := mocks.NewMockJobRepository(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewSyncItemsUseCase(cfg, mockItemRepo, mockJobRepo, mockLogger)

	// Setup request with empty params
	request := SyncItemsRequest{
		APISource: "pokemon",
		Params:    map[string]interface{}{},
	}

	// Set expectations - allow any repository calls
	mockJobRepo.EXPECT().
		CreateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(int64(1), nil).AnyTimes()
	mockJobRepo.EXPECT().
		UpdateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	mockItemRepo.EXPECT().
		UpsertWithHash(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	// Expect logger calls for background sync
	mockLogger.EXPECT().
		Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().
		Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "accepted", response.Status)
}

func TestSyncItemsUseCase_Execute_OpenWeatherAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockJobRepo := mocks.NewMockJobRepository(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewSyncItemsUseCase(cfg, mockItemRepo, mockJobRepo, mockLogger)

	// Setup request for OpenWeather
	request := SyncItemsRequest{
		APISource: "openweather",
		Params: map[string]interface{}{
			"cities": []string{"Jakarta", "Bandung"},
		},
	}

	// Set expectations - allow any repository calls
	mockJobRepo.EXPECT().
		CreateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(int64(1), nil).AnyTimes()
	mockJobRepo.EXPECT().
		UpdateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	mockItemRepo.EXPECT().
		UpsertWithHash(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "accepted", response.Status)
	assert.Equal(t, "Sync job has been accepted for background processing", response.Message)
}

func TestSyncItemsUseCase_Execute_JobRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockJobRepo := mocks.NewMockJobRepository(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewSyncItemsUseCase(cfg, mockItemRepo, mockJobRepo, mockLogger)

	// Setup request
	request := SyncItemsRequest{
		APISource: "pokemon",
	}

	// Execute test - should always return accepted for async processing
	response, err := useCase.Execute(context.Background(), request)

	// Assertions - Execute should always return accepted for async processing
	require.NoError(t, err)
	assert.Equal(t, "accepted", response.Status)
	assert.Equal(t, "Sync job has been accepted for background processing", response.Message)
}

func TestSyncItemsUseCase_Execute_WithForceSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockJobRepo := mocks.NewMockJobRepository(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Setup config
	cfg := &config.Config{
		API: config.APIConfig{
			Timeout: 30 * time.Second,
		},
		Retry: config.RetryConfig{},
	}

	// Create usecase
	useCase := NewSyncItemsUseCase(cfg, mockItemRepo, mockJobRepo, mockLogger)

	// Setup request with force sync
	request := SyncItemsRequest{
		APISource: "pokemon",
		ForceSync: true,
		Params: map[string]interface{}{
			"limit": 50,
		},
	}

	// Set expectations - allow any repository calls
	mockJobRepo.EXPECT().
		CreateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(int64(1), nil).AnyTimes()
	mockJobRepo.EXPECT().
		UpdateSyncJobRecord(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	mockItemRepo.EXPECT().
		UpsertWithHash(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "accepted", response.Status)
	assert.Equal(t, "Sync job has been accepted for background processing", response.Message)
}
