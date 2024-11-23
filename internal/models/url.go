package models

type StorageURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UUID        int    `json:"uuid"`
}
