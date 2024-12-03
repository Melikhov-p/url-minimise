package repository

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type Storage interface {
	AddURL(context.Context, *models.StorageURL) (string, error)
	AddURLs(context.Context, []*models.StorageURL) error
	GetFullURL(context.Context, string) (string, error)
	GetShortURL(context.Context, *sql.Tx, string) (string, error)
	CheckShort(context.Context, string) bool
	Ping(context.Context) error
}

type StorageSaver interface { // Для хранилищ, которым нужен отдельный метод для сохранения данных, например файл
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

		ctx := context.Background()
		if err = store.Ping(ctx); err != nil {
			return nil, fmt.Errorf("error ping database %w", err)
		}

		if err = makeMigrations(cfg, store.DB); err != nil {
			return nil, fmt.Errorf("error making migrations %w", err)
		}

		return store, nil
	}

	return nil, fmt.Errorf("unknow type of store %d", cfg.StorageMode)
}

func makeMigrations(cfg *config.Config, db *sql.DB) error {
	var err error

	if err = goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("error goose set dialect %w", err)
	}

	if err = goose.Up(db, cfg.Storage.Database.MigrationsPath); err != nil {
		return fmt.Errorf("error up migrations %w", err)
	}

	return nil
}
