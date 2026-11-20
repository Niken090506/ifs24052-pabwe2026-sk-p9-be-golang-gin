package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	os.Unsetenv("APP_PORT")
	cfg := LoadConfig()
	assert.Equal(t, "8081", cfg.AppPort)

	os.Setenv("APP_PORT", "9090")
	defer os.Unsetenv("APP_PORT")
	cfg2 := LoadConfig()
	assert.Equal(t, "9090", cfg2.AppPort)
}
