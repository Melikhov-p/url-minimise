package repository

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage interface {
	AddURL(*models.StorageURL)
	GetFullURL(string) string
	CheckEl(string) bool
	Ping() error
}

type StorageSaver interface {
	Save(*models.StorageURL) error
}

func NewStorage(cfg *config.Config) (Storage, error) {
	switch cfg.StorageMode {
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
	case storage.StorageInDatabase:
		db, err := sql.Open("pgx", cfg.Storage.Database.DSN)

		if err != nil {
			return nil, fmt.Errorf("error open conn with pgx: ERROR: %w, DSN: %s", err, cfg.Storage.Database.DSN)
		}

		store := &storage.DatabaseStorage{
			DB: db,
		}

		return store, nil
	}

	return nil, fmt.Errorf("unknow type of store %d", cfg.StorageMode)
}
