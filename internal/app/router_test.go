package app

import (
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreateRouter(t *testing.T) {
	log, err := logger.BuildLogger("DEBUG")
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)
	store, err := repository.NewStorage(cfg, log)
	assert.NoError(t, err)

	r := CreateRouter(cfg, store, log)
	assert.IsType(t, chi.NewRouter(), r)
}
