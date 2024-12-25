package service

import (
	"context"
	"fmt"

	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
)

func GetDeleteTasksWStatus(
	ctx context.Context,
	status models.DelTaskStatus,
	storage repository.Storage,
) ([]*models.DelTask, error) {
	tasks, err := storage.GetDeleteTasksWStatus(ctx, status)
	if err != nil {
		return tasks, fmt.Errorf("error getting tasks %w", err)
	}
	return tasks, nil
}

func UpdateTasksStatus(
	ctx context.Context,
	tasks []*models.DelTask,
	newStatus models.DelTaskStatus,
	storage repository.Storage,
) error {
	err := storage.UpdateTasksStatus(ctx, tasks, newStatus)
	if err != nil {
		return fmt.Errorf("error updating statuses for tasks %w", err)
	}

	return nil
}
