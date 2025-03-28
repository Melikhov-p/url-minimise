package repository

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Melikhov-p/url-minimise/internal/auth"
	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver.
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

// Storage интерфейс хранилища.
type Storage interface {
	AddURL(context.Context, *models.StorageURL) (string, error)
	AddURLs(context.Context, []*models.StorageURL) error
	GetDeleteTasksWStatus(ctx context.Context, status models.DelTaskStatus) ([]*models.DelTask, error)
	MarkAsDeletedURL(ctx context.Context, tasks []*models.DelTask) error
	UpdateTasksStatus(ctx context.Context, tasks []*models.DelTask, newStatus models.DelTaskStatus) error
	AddDeleteTask(shortURL []string, userID int) error
	GetURL(context.Context, string) (*models.StorageURL, error)
	GetShortURL(context.Context, *sql.Tx, string) (string, error)
	CheckShort(context.Context, string) bool
	Ping(context.Context) error
	AddUser(ctx context.Context) (*models.User, error)
	GetURLsByUserID(ctx context.Context, userID int) ([]*models.StorageURL, error)
	Close() error
	GetURLsCount(ctx context.Context) (int, error)
	GetUsersCount(ctx context.Context) (int, error)
}

// StorageSaver для хранилищ, которым нужен отдельный метод для сохранения данных, например файл.
type StorageSaver interface { // Для хранилищ, которым нужен отдельный метод для сохранения данных, например файл.
	Save(*models.StorageURL) error
}

// NewStorage возвращает объект хранилища.
// Один из: in_memory | file | database.
func NewStorage(cfg *config.Config, _ *zap.Logger) (Storage, error) {
	switch cfg.StorageMode {
	case storage.BaseStorage:
		key, err := auth.GenerateAuthKey()
		if err != nil {
			return nil, fmt.Errorf("error generating secret key for storage %w", err)
		}
		cfg.SecretKey = key
		return storage.NewMemoryStorage(), nil
	case storage.StorageFromFile:
		file, err := os.OpenFile(cfg.Storage.FileStorage.FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			return nil, fmt.Errorf("error opening storage file %w", err)
		}

		scan := bufio.NewScanner(file)

		store := &storage.FileStorage{
			MemoryStorage: *storage.NewMemoryStorage(),
			File:          file,
			Encoder:       json.NewEncoder(file),
		}

		var element models.StorageURL
		for scan.Scan() {
			err = json.Unmarshal(scan.Bytes(), &element)
			if err != nil {
				return nil, fmt.Errorf("error unmarshal url from model %w", err)
			}
			store.SetInMemory(element.ShortURL, &element)
		}

		key, err := auth.GenerateAuthKey()
		if err != nil {
			return nil, fmt.Errorf("error generating secret key for storage %w", err)
		}
		cfg.SecretKey = key
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

		key, err := store.GetSecretKey(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting secret key for database %w", err)
		}
		cfg.SecretKey = key

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
