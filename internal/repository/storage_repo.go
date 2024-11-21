package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/storage"
)

type IStorage interface {
	AddURL(*models.StorageURL)
	GetFullURL(string) string
	GetDB() map[string]*models.StorageURL
}

type IStorageSaver interface {
	Save(*models.StorageURL) error
}

func NewStorage(storageType storage.StorageType, cfg *config.Config) (IStorage, error) {
	switch storageType {
	case storage.BaseStorage:
		return &storage.MemoryStorage{
			DB: map[string]*models.StorageURL{},
		}, nil
	case storage.StorageFromFile:
		file, err := os.OpenFile(cfg.Storage.FileStorage.FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			return nil, fmt.Errorf("error opening storage file %w", err)
		}

		scan := bufio.NewScanner(file)

		store := &storage.FileStorage{
			MemoryStorage: storage.MemoryStorage{
				DB: map[string]*models.StorageURL{},
			},
			File:    file,
			Encoder: json.NewEncoder(file),
		}

		var element models.StorageURL
		for scan.Scan() {
			err = json.Unmarshal(scan.Bytes(), &element)
			if err != nil {
				return nil, fmt.Errorf("error unmarshal url from model %w", err)
			}
			store.DB[element.ShortURL] = &element
		}

		return store, nil
	}

	return nil, fmt.Errorf("unknow type of store %d", storageType)
}
