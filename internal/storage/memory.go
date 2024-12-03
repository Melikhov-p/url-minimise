package storage

import (
	"context"
	"database/sql"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

type MemoryStorage struct {
	DB map[string]*models.StorageURL
}

func (s *MemoryStorage) AddURL(ctx context.Context, newURL *models.StorageURL) (string, error) {
	if short, ok := s.checkFull(ctx, newURL.OriginalURL); ok {
		return short, ErrOriginalURLExist
	}
	s.DB[newURL.ShortURL] = newURL
	return newURL.ShortURL, nil
}

func (s *MemoryStorage) AddURLs(_ context.Context, newURLs []*models.StorageURL) error {
	for _, url := range newURLs {
		s.DB[url.ShortURL] = url
	}

	return nil
}

func (s *MemoryStorage) GetShortURL(_ context.Context, _ *sql.Tx, fullURL string) (string, error) {
	var short string

	for _, url := range s.DB {
		if url.OriginalURL == fullURL {
			short = url.ShortURL
			return short, nil
		}
	}

	return "", ErrNotFound
}

func (s *MemoryStorage) GetFullURL(_ context.Context, shortURL string) (string, error) {
	searchedElem := s.DB[shortURL]
	if searchedElem != nil {
		return searchedElem.OriginalURL, nil
	}
	return "", ErrNotFound
}

func (s *MemoryStorage) CheckShort(_ context.Context, short string) bool { return s.DB[short] != nil }

// Если оригинальный URL есть в базе - true.
func (s *MemoryStorage) checkFull(_ context.Context, fullURL string) (string, bool) {
	for _, url := range s.DB {
		if fullURL == url.OriginalURL {
			return url.ShortURL, true
		}
	}

	return "", false
}

func (s *MemoryStorage) Ping(_ context.Context) error { return nil }
