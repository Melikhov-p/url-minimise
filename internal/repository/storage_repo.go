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
)

const dbTableName = "url"

type Storage interface {
	AddURL(context.Context, *models.StorageURL) error
	AddURLs(context.Context, []*models.StorageURL) error
	GetFullURL(context.Context, string) (string, error)
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

		ok, err := dbCheckTable(ctx, store.DB, dbTableName)
		if err != nil {
			return nil, fmt.Errorf("error check table in database %w", err)
		}

		if !ok {
			if err = createTable(ctx, store.DB); err != nil {
				return nil, fmt.Errorf("error creating table %w", err)
			}
		}

		return store, nil
	}

	return nil, fmt.Errorf("unknow type of store %d", cfg.StorageMode)
}

func dbCheckTable(ctx context.Context, db *sql.DB, tableName string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT FROM
                information_schema.tables
            WHERE
                table_schema = 'public' AND
                table_name = $1
        )
    `
	var exist bool
	rows := db.QueryRowContext(ctx, query, tableName)

	if err := rows.Scan(&exist); err != nil {
		return false, fmt.Errorf("error scanning row from database %w", err)
	}
	return exist, nil
}

func createTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE URL (
			short_url VARCHAR(255) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL
		);`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("error exec context from database %w", err)
	}

	return nil
}
