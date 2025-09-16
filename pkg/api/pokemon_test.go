package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPokemonClient_extractPokemonID(t *testing.T) {
	client := &PokemonClient{}

	tests := []struct {
		name string
		url  string
		want int
	}{
		{
			name: "extract ID from pokemon URL",
			url:  "https://pokeapi.co/api/v2/pokemon/1/",
			want: 1,
		},
		{
			name: "extract ID from pokemon URL without trailing slash",
			url:  "https://pokeapi.co/api/v2/pokemon/150",
			want: 150,
		},
		{
			name: "invalid URL",
			url:  "https://pokeapi.co/api/v2/berry/1/",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.extractPokemonID(tt.url)
			assert.Equal(t, tt.want, got, "extractPokemonID should return correct ID")
		})
	}
}

func TestPokemonClient_transformPokemonResponse(t *testing.T) {
	client := &PokemonClient{}

	// Sample Pokemon API response
	sampleResponse := `{
		"count": 1302,
		"next": "https://pokeapi.co/api/v2/pokemon?offset=20&limit=20",
		"previous": null,
		"results": [
			{"name": "bulbasaur", "url": "https://pokeapi.co/api/v2/pokemon/1/"},
			{"name": "ivysaur", "url": "https://pokeapi.co/api/v2/pokemon/2/"},
			{"name": "venusaur", "url": "https://pokeapi.co/api/v2/pokemon/3/"}
		]
	}`

	var response PokemonResponse
	err := json.Unmarshal([]byte(sampleResponse), &response)
	assert.NoError(t, err, "Should unmarshal sample response without error")

	items := client.transformPokemonResponse(response)

	assert.Len(t, items, 3, "Should return 3 items")

	// Test first item
	assert.Equal(t, 1, items[0].ID, "First item should have ID 1")
	assert.Equal(t, "bulbasaur", items[0].Title, "First item should have title 'bulbasaur'")
	assert.Equal(t, "pokemon", items[0].ExtendInfo["api_source"], "Should set api_source to 'pokemon'")
	assert.Equal(t, "https://pokeapi.co/api/v2/pokemon/1/", items[0].ExtendInfo["url"], "Should store original URL")
	assert.NotNil(t, items[0].ExtendInfo["raw_data"], "Should store raw data")

	// Test second item
	assert.Equal(t, 2, items[1].ID, "Second item should have ID 2")
	assert.Equal(t, "ivysaur", items[1].Title, "Second item should have title 'ivysaur'")

	// Test third item
	assert.Equal(t, 3, items[2].ID, "Third item should have ID 3")
	assert.Equal(t, "venusaur", items[2].Title, "Third item should have title 'venusaur'")
}
