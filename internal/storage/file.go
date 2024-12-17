package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

type FileStorage struct {
	MemoryStorage
	File    *os.File
	Encoder *json.Encoder
	Scanner *bufio.Scanner
}

// SetInMemory Жесткая установка связки в in-memory хранилище при загрузке данных из файла.
func (s *FileStorage) SetInMemory(shortURL string, newURL *models.StorageURL) {
	s.urls[shortURL] = newURL
}

func (s *FileStorage) Save(record *models.StorageURL) error {
	if err := s.Encoder.Encode(record); err != nil {
		return fmt.Errorf("error encoding json to model %w", err)
	}
	return nil
}
