package storage

import (
	"github.com/Melikhov-p/url-minimise/internal/models"
)

type MemoryStorage struct {
	DB map[string]*models.StorageURL
}

func (s *MemoryStorage) AddURL(newURL *models.StorageURL) {
	s.DB[newURL.ShortURL] = newURL
}

func (s *MemoryStorage) GetFullURL(shortURL string) string {
	searchedElem := s.DB[shortURL]
	if searchedElem != nil {
		return searchedElem.OriginalURL
	}
	return ""
}

func (s *MemoryStorage) GetDB() map[string]*models.StorageURL {
	return s.DB
}
