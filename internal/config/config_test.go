package config

import (
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	log, err := logger.BuildLogger("DEBUG")
	assert.NoError(t, err)
	cfg := NewConfig(log, true)
	assert.Equal(t, defaultSrvAddr, cfg.ServerAddr)
}

func TestBuild(t *testing.T) {
	log, err := logger.BuildLogger("DEBUG")
	assert.NoError(t, err)
	cfg := NewConfig(log, true)
	assert.Equal(t, defaultSrvAddr, cfg.ServerAddr)
	cfg.build(log)
}
