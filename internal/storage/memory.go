package storage

import (
	"context"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

type MemoryStorage struct {
	DB map[string]*models.StorageURL
}

func (s *MemoryStorage) AddURL(_ context.Context, newURL *models.StorageURL) error {
	s.DB[newURL.ShortURL] = newURL
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

func (s *MemoryStorage) Ping(_ context.Context) error { return nil }
