package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/internal/item/usecase/mocks"
	loggermocks "github.com/zainokta/item-sync/pkg/logger/mocks"
)

func TestListItemsUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Create usecase
	useCase := NewListItemsUseCase(mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := ListItemsRequest{
		Limit:     10,
		Offset:    0,
		APISource: "pokemon",
	}

	// Mock data
	mockItems := []entity.Item{
		{ID: 1, Title: "Pikachu", APISource: "pokemon"},
		{ID: 2, Title: "Charizard", APISource: "pokemon"},
	}

	// Set expectations
	mockCache.EXPECT().
		GetItems(gomock.Any(), "items:pokemon::10:0").
		Return([]entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByAPISource(gomock.Any(), "pokemon", 10, 0).
		Return(mockItems, nil)
	mockCache.EXPECT().
		SetItems(gomock.Any(), "items:pokemon::10:0", mockItems, 10*time.Minute).
		Return(nil)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, response.Items, 2)
	assert.Equal(t, 2, response.TotalCount)
	assert.Equal(t, "Pikachu", response.Items[0].Title)
	assert.Equal(t, "Charizard", response.Items[1].Title)
}

func TestListItemsUseCase_Execute_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Create usecase
	useCase := NewListItemsUseCase(mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := ListItemsRequest{
		Limit:  10,
		Offset: 0,
		Status: "active",
	}

	// Mock data
	mockItems := []entity.Item{
		{ID: 1, Title: "Item 1", APISource: "test"},
		{ID: 2, Title: "Item 2", APISource: "test"},
	}

	// Set expectations - cache hit
	mockCache.EXPECT().
		GetItems(gomock.Any(), "items::active:10:0").
		Return(mockItems, nil)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, response.Items, 2)
	assert.Equal(t, 0, response.TotalCount) // Cache doesn't store total count, returns len(items)
}

func TestListItemsUseCase_Execute_DefaultValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Create usecase
	useCase := NewListItemsUseCase(mockItemRepo, mockCache, mockLogger)

	// Setup request with invalid values
	request := ListItemsRequest{
		Limit:  -1, // Should default to 20
		Offset: -5, // Should default to 0
	}

	// Mock data
	mockItems := []entity.Item{
		{ID: 1, Title: "Default Item"},
	}

	// Set expectations
	mockCache.EXPECT().
		GetItems(gomock.Any(), "items:::20:0").
		Return([]entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindAll(gomock.Any(), 20, 0).
		Return(mockItems, nil)
	mockCache.EXPECT().
		SetItems(gomock.Any(), "items:::20:0", mockItems, 10*time.Minute).
		Return(nil)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, 1, response.TotalCount)
}

func TestListItemsUseCase_Execute_FindByStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Create usecase
	useCase := NewListItemsUseCase(mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := ListItemsRequest{
		Limit:  5,
		Offset: 10,
		Status: "completed",
	}

	// Mock data
	mockItems := []entity.Item{
		{ID: 1, Title: "Completed Item 1", APISource: "test"},
	}

	// Set expectations
	mockCache.EXPECT().
		GetItems(gomock.Any(), "items::completed:5:10").
		Return([]entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByStatus(gomock.Any(), "completed", 5, 10).
		Return(mockItems, nil)
	mockCache.EXPECT().
		SetItems(gomock.Any(), "items::completed:5:10", mockItems, 10*time.Minute).
		Return(nil)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, "test", response.Items[0].APISource)
}

func TestListItemsUseCase_Execute_FindByType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Create usecase
	useCase := NewListItemsUseCase(mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := ListItemsRequest{
		Limit:    15,
		Offset:   5,
		ItemType: "pokemon",
	}

	// Mock data
	mockItems := []entity.Item{
		{ID: 1, Title: "Pokemon Item", APISource: "pokemon"},
	}

	// Set expectations
	mockCache.EXPECT().
		GetItems(gomock.Any(), "items:::15:5").
		Return([]entity.Item{}, assert.AnError) // Cache miss (ItemType not in cache key)
	mockItemRepo.EXPECT().
		FindByType(gomock.Any(), "pokemon", 15, 5).
		Return(mockItems, nil)
	mockCache.EXPECT().
		SetItems(gomock.Any(), "items:::15:5", mockItems, 10*time.Minute).
		Return(nil)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, "pokemon", response.Items[0].APISource)
}

func TestListItemsUseCase_Execute_EmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Create usecase
	useCase := NewListItemsUseCase(mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := ListItemsRequest{
		Limit:  10,
		Offset: 0,
		Status: "nonexistent",
	}

	// Set expectations
	mockCache.EXPECT().
		GetItems(gomock.Any(), "items::nonexistent:10:0").
		Return([]entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByStatus(gomock.Any(), "nonexistent", 10, 0).
		Return([]entity.Item{}, nil)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.NoError(t, err)
	assert.Empty(t, response.Items)
	assert.Equal(t, 0, response.TotalCount)
}

func TestListItemsUseCase_Execute_CacheSetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Create usecase
	useCase := NewListItemsUseCase(mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := ListItemsRequest{
		Limit:  10,
		Offset: 0,
	}

	// Mock data
	mockItems := []entity.Item{
		{ID: 1, Title: "Test Item"},
	}

	// Set expectations
	mockCache.EXPECT().
		GetItems(gomock.Any(), "items:::10:0").
		Return([]entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindAll(gomock.Any(), 10, 0).
		Return(mockItems, nil)
	mockCache.EXPECT().
		SetItems(gomock.Any(), "items:::10:0", mockItems, 10*time.Minute).
		Return(assert.AnError)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions - should still succeed even if cache fails
	require.NoError(t, err)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, 1, response.TotalCount)
}

func TestListItemsUseCase_Execute_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockItemRepo := mocks.NewMockItemRepository(ctrl)
	mockCache := mocks.NewMockItemCache(ctrl)
	mockLogger := loggermocks.NewMockLogger(ctrl)

	// Create usecase
	useCase := NewListItemsUseCase(mockItemRepo, mockCache, mockLogger)

	// Setup request
	request := ListItemsRequest{
		Limit:     10,
		Offset:    0,
		APISource: "pokemon",
	}

	// Set expectations
	mockCache.EXPECT().
		GetItems(gomock.Any(), "items:pokemon::10:0").
		Return([]entity.Item{}, assert.AnError) // Cache miss
	mockItemRepo.EXPECT().
		FindByAPISource(gomock.Any(), "pokemon", 10, 0).
		Return([]entity.Item{}, assert.AnError)

	// Execute test
	response, err := useCase.Execute(context.Background(), request)

	// Assertions
	require.Error(t, err)
	assert.Equal(t, ListItemsResponse{}, response)
}