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

func TestGetDeleteTasksWStatus(t *testing.T) {
	log, err := logger.BuildLogger("ERROR")
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)
	store, err := repository.NewStorage(cfg, log)
	assert.NoError(t, err)

	_, err = GetDeleteTasksWStatus(
		context.Background(),
		models.Registered,
		store)
	assert.NoError(t, err)
}

func TestUpdateTasksStatus(t *testing.T) {
	log, err := logger.BuildLogger("ERROR")
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)
	store, err := repository.NewStorage(cfg, log)
	assert.NoError(t, err)

	task := &models.DelTask{
		URL:    "short",
		UserID: 1,
		Status: models.Registered,
	}
	var taskList []*models.DelTask
	taskList = append(taskList, task)

	err = UpdateTasksStatus(context.Background(), taskList, models.Done, store)
	assert.NoError(t, err)
}
