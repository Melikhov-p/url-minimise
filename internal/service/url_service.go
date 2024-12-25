package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	storagePkg "github.com/Melikhov-p/url-minimise/internal/storage"
	"go.uber.org/zap"
)

// AddURL Adding new url in storage, return URL model and error.
func AddURL(
	ctx context.Context,
	storage repository.Storage,
	logger *zap.Logger,
	originalURL string,
	cfg *config.Config,
	userID int,
) (*models.StorageURL, error) {
	newURL, err := repository.NewStorageURL(ctx, originalURL, storage, cfg, userID)
	if err != nil {
		logger.Error("error creating short URL", zap.Error(err))
		return nil, fmt.Errorf("error creating short URL model %w", err)
	}

	if short, err := storage.AddURL(ctx, newURL); err != nil {
		if errors.Is(err, storagePkg.ErrOriginalURLExist) {
			logger.Info("original URL already exist in storage", zap.String("OriginalURL", newURL.OriginalURL))
			newURL.ShortURL = short
			return newURL, storagePkg.ErrOriginalURLExist
		}
		logger.Error("error adding new url", zap.Error(err))
		return nil, fmt.Errorf("error adding new URL %w", err)
	}

	return newURL, nil
}

func MarkAsDeleted(
	ctx context.Context,
	tasks []*models.DelTask,
	storage repository.Storage,
) error {
	err := storage.MarkAsDeletedURL(ctx, tasks)
	if err != nil {
		return fmt.Errorf("error mark URL deleted %w", err)
	}

	return nil
}
