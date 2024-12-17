package models

type StorageURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UUID        string `json:"uuid"`
	UserID      int    `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}
