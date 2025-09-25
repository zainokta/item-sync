package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		expectedType int
	}{
		{
			name:         "default Max Idle conns should be pokemon",
			envValue:     "",
			expectedType: 10,
		},
		{
			name:         "can set Max Idle conns to custom value",
			envValue:     "50",
			expectedType: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv("API_MAX_IDLE_CONNS")

			if tt.envValue != "" {
				os.Setenv("API_MAX_IDLE_CONNS", tt.envValue)
				defer os.Unsetenv("API_MAX_IDLE_CONNS")
			}

			cfg, err := LoadConfig()
			assert.NoError(t, err, "LoadConfig should not return an error")
			assert.Equal(t, tt.expectedType, cfg.API.MaxIdleConns, "Max Idle conns should match expected value")
		})
	}
}
