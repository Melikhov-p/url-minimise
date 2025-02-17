package service

import (
	"context"
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestAddURL(t *testing.T) {
	log, err := logger.BuildLogger("ERROR")
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)
	store, err := repository.NewStorage(cfg, log)
	assert.NoError(t, err)

	_, err = AddURL(context.Background(), store, log, "original", cfg, 1)
	assert.NoError(t, err)
}

func TestMarkAsDeleted(t *testing.T) {
	log, err := logger.BuildLogger("ERROR")
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)
	store, err := repository.NewStorage(cfg, log)
	assert.NoError(t, err)

	short, err := AddURL(context.Background(), store, log, "original", cfg, 1)
	assert.NoError(t, err)

	task := &models.DelTask{
		URL:    short.ShortURL,
		UserID: 1,
		Status: models.Registered,
	}
	var taskList []*models.DelTask
	taskList = append(taskList, task)
	err = MarkAsDeleted(context.Background(), taskList, store)
	assert.NoError(t, err)
}
