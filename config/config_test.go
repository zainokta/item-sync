package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIConfig_APIType(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		expectedType string
	}{
		{
			name:         "default API type should be pokemon",
			envValue:     "",
			expectedType: "pokemon",
		},
		{
			name:         "can set API type to openweather",
			envValue:     "openweather",
			expectedType: "openweather",
		},
		{
			name:         "can set API type to custom value",
			envValue:     "custom",
			expectedType: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv("API_API_TYPE")

			if tt.envValue != "" {
				os.Setenv("API_API_TYPE", tt.envValue)
				defer os.Unsetenv("API_API_TYPE")
			}

			cfg, err := LoadConfig()
			assert.NoError(t, err, "LoadConfig should not return an error")
			assert.Equal(t, tt.expectedType, cfg.API.APIType, "API type should match expected value")
		})
	}
}
