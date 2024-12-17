package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

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
	storage repository.Storage,
	logger *zap.Logger,
	shortURLs []string,
	user *models.User,
	_ *config.Config,
) {
	inCh := generator(user, shortURLs...)
	ch1 := storage.MarkAsDeletedURL(ctx, inCh)
	ch2 := storage.MarkAsDeletedURL(ctx, inCh)

	for res := range fanIn(ch1, ch2) {
		logger.Info("request delete URL",
			zap.Bool("result", res.Res),
			zap.Error(res.Err),
			zap.String("URL", res.URL))
	}
}

func generator(user *models.User, urls ...string) chan storagePkg.MarkDeleteURL {
	outCh := make(chan storagePkg.MarkDeleteURL)
	go func() {
		defer close(outCh)
		for _, url := range urls {
			outCh <- storagePkg.MarkDeleteURL{
				ShortURL: url,
				User:     user,
			}
		}
	}()

	return outCh
}

func fanIn(chans ...chan storagePkg.MarkDeleteResult) chan storagePkg.MarkDeleteResult {
	resultCh := make(chan storagePkg.MarkDeleteResult)

	wg := sync.WaitGroup{}

	for _, ch := range chans {
		innerCh := ch
		wg.Add(1)

		go func() {
			defer wg.Done()

			for x := range innerCh {
				resultCh <- x
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh
}
