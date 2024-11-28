package storage

import (
	"context"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

type MemoryStorage struct {
	DB map[string]*models.StorageURL
}

func (s *MemoryStorage) AddURL(ctx context.Context, newURL *models.StorageURL) error {
	if s.checkFull(ctx, newURL.OriginalURL) {
		return ErrURLExist
	}
	s.DB[newURL.ShortURL] = newURL
	return nil
}

func (s *MemoryStorage) AddURLs(ctx context.Context, newURLs []*models.StorageURL) error {
	for _, url := range newURLs {
		s.DB[url.ShortURL] = url
	}

	return nil
}

func (s *MemoryStorage) GetFullURL(_ context.Context, shortURL string) (string, error) {
	searchedElem := s.DB[shortURL]
	if searchedElem != nil {
		return searchedElem.OriginalURL, nil
	}
	return "", errNotFound
}

func (s *MemoryStorage) CheckShort(_ context.Context, short string) bool { return s.DB[short] != nil }

// Если оригинальный URL есть в базе - true.
func (s *MemoryStorage) checkFull(_ context.Context, fullURL string) bool {
	for _, url := range s.DB {
		if fullURL == url.OriginalURL {
			return true
		}
	}

	return false
}

func (s *MemoryStorage) Ping(_ context.Context) error { return nil }
