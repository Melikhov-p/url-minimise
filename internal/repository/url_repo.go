package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
)

func NewStorageURL(ctx context.Context,
	fullURL string, s Storage,
	cfg *config.Config,
	userID int) (*models.StorageURL, error) {
	short, err := randomString(ctx, cfg.ShortURLSize, s)

	if err == nil {
		return &models.StorageURL{
			ShortURL:    short,
			OriginalURL: fullURL,
			UserID:      userID,
		}, nil
	}
	return nil, err
}

func NewStorageMultiURL(
	ctx context.Context,
	fullURLs []string,
	s Storage,
	cfg *config.Config,
	userID int) ([]*models.StorageURL, error) {
	newURLs := make([]*models.StorageURL, 0, len(fullURLs))

	for _, url := range fullURLs {
		short, err := randomString(ctx, cfg.ShortURLSize, s)
		if err != nil {
			return nil, fmt.Errorf("error creating random string %w", err)
		}
		newURLs = append(newURLs, &models.StorageURL{
			ShortURL:    short,
			OriginalURL: url,
			UserID:      userID,
		})
	}

	return newURLs, nil
}

// Создает рандомную строку заданного размера.
func randomString(ctx context.Context, size int, s Storage) (string, error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	tries := 5 // количество попыток создать уникальную строку

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	for tries > 0 {
		b := make([]rune, size)
		for i := range b {
			b[i] = chars[rnd.Intn(len(chars))]
		}
		str := string(b)
		time.Sleep(1 * time.Nanosecond) // Иначе в batch есть шанс два одинаковых сгенерить
		if ok := s.CheckShort(ctx, str); !ok {
			return str, nil
		}
		tries--
	}

	return "", errors.New("reached max tries limit")
}
