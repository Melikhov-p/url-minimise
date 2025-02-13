package models

import "sync"

// StorageURL структура хранимого в хранилище URL
type StorageURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UUID        string `json:"uuid"`
	UserID      int    `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

// DelURLs адрес отмеченные на удаление
type DelURLs struct {
	URLs []string
	Mu   sync.Mutex
}
