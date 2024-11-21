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

func (s *FileStorage) Save(record *models.StorageURL) error {
	if err := s.Encoder.Encode(record); err != nil {
		return fmt.Errorf("error encoding json to model %w", err)
	}
	return nil
}
