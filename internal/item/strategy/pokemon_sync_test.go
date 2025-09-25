package strategy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/pkg/api"
)

type mockPokemonAPIClient struct {
	pages [][]entity.ExternalItem
	calls int
}

func (m *mockPokemonAPIClient) Fetch(ctx context.Context, apiName string, operation string, params map[string]interface{}) ([]entity.ExternalItem, error) {
	if m.calls >= len(m.pages) {
		return []entity.ExternalItem{}, nil
	}

	page := m.pages[m.calls]
	m.calls++

	return page, nil
}

func (m *mockPokemonAPIClient) FetchPaginated(ctx context.Context, apiName string, operation string, params map[string]interface{}) (*api.PaginatedResponse, error) {
	if m.calls >= len(m.pages) {
		return &api.PaginatedResponse{
			Items:      []entity.ExternalItem{},
			Pagination: api.NewPaginationMetadata(0, "", ""),
		}, nil
	}

	page := m.pages[m.calls]
	m.calls++

	var pagination *api.PaginationMetadata
	if m.calls < len(m.pages) {
		// Has next page
		nextURL := "https://pokeapi.co/api/v2/pokemon?offset=20&limit=2"
		pagination = api.NewPaginationMetadata(1000, nextURL, "")
	} else {
		// Last page
		pagination = api.NewPaginationMetadata(1000, "", "")
	}

	return &api.PaginatedResponse{
		Items:      page,
		Pagination: pagination,
	}, nil
}

func (m *mockPokemonAPIClient) FetchByID(ctx context.Context, apiName string, id int) (entity.ExternalItem, error) {
	return entity.ExternalItem{}, nil
}

func TestPokemonSyncStrategy_FetchAllItems(t *testing.T) {
	mockClient := &mockPokemonAPIClient{
		pages: [][]entity.ExternalItem{
			// Page 1
			{
				{ID: 1, Title: "bulbasaur", ExtendInfo: make(map[string]interface{})},
				{ID: 2, Title: "ivysaur", ExtendInfo: make(map[string]interface{})},
			},
			// Page 2
			{
				{ID: 3, Title: "venusaur", ExtendInfo: make(map[string]interface{})},
				{ID: 4, Title: "charmander", ExtendInfo: make(map[string]interface{})},
			},
			// Page 3 (last page)
			{
				{ID: 5, Title: "charmeleon", ExtendInfo: make(map[string]interface{})},
			},
		},
	}

	strategy := NewPokemonSyncStrategy(nil, mockClient)
	request := SyncItemsRequest{
		APISource: "pokemon",
		Operation: "list",
		Params:    map[string]interface{}{"limit": 2},
	}

	items, err := strategy.FetchAllItems(context.Background(), request)

	assert.NoError(t, err, "Should fetch all items without error")
	assert.Len(t, items, 5, "Should fetch all items across all pages")
	assert.Equal(t, 3, mockClient.calls, "Should make 3 API calls for 3 pages")

	// Verify items are in correct order
	expectedTitles := []string{"bulbasaur", "ivysaur", "venusaur", "charmander", "charmeleon"}
	for i, item := range items {
		assert.Equal(t, expectedTitles[i], item.Title, "Items should be in correct order")
		assert.Equal(t, i+1, item.ID, "Items should have correct IDs")
	}
}

func TestPokemonSyncStrategy_ParseNextURL(t *testing.T) {
	strategy := NewPokemonSyncStrategy(nil, nil)

	tests := []struct {
		name       string
		url        string
		wantOffset int
		wantLimit  int
	}{
		{
			name:       "valid next URL",
			url:        "https://pokeapi.co/api/v2/pokemon?offset=20&limit=20",
			wantOffset: 20,
			wantLimit:  20,
		},
		{
			name:       "URL without limit defaults to 20",
			url:        "https://pokeapi.co/api/v2/pokemon?offset=40",
			wantOffset: 40,
			wantLimit:  20,
		},
		{
			name:       "URL without offset defaults to 0",
			url:        "https://pokeapi.co/api/v2/pokemon?limit=10",
			wantOffset: 0,
			wantLimit:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, limit, err := strategy.parseNextURL(tt.url)

			assert.NoError(t, err, "Should parse valid URL without error")

			assert.Equal(t, tt.wantOffset, offset, "Should parse correct offset")
			assert.Equal(t, tt.wantLimit, limit, "Should parse correct limit")
		})
	}
}
