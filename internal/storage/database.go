package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

type DatabaseStorage struct {
	DB *sql.DB
}

const dbTimeout = 15 * time.Second

func (db *DatabaseStorage) AddURL(ctx context.Context, newURL *models.StorageURL) error {
	query := `
		INSERT INTO URL (short_url, original_url)
        VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	_, err := db.DB.ExecContext(ctx, query, newURL.ShortURL, newURL.OriginalURL)
	if err != nil {
		return fmt.Errorf("error exec context from database in addurl %w", err)
	}

	return nil
}

func (db *DatabaseStorage) AddURLs(ctx context.Context, newURLs []*models.StorageURL) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("error begin transaction %w", err)
	}
	// если Commit будет раньше, то откат проигнорируется
	defer func() {
		_ = tx.Rollback()
	}()

	preparedQuery, err := tx.PrepareContext(ctx, `
		INSERT INTO URL (short_url, original_url)
        VALUES ($1, $2)
	`)
	if err != nil {
		return fmt.Errorf("error getting prepared query %w", err)
	}
	defer func() {
		_ = preparedQuery.Close()
	}()

	for _, url := range newURLs {
		if _, err = preparedQuery.ExecContext(ctx, url.ShortURL, url.OriginalURL); err != nil {
			return fmt.Errorf("error executing context for query in addURLs %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error commiting changes in database %w", err)
	}

	return nil
}

func (db *DatabaseStorage) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	query := `
		SELECT original_url FROM URL WHERE short_url = $1
	`

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var fullURL string
	if err := db.DB.QueryRowContext(ctx, query, shortURL).Scan(&fullURL); err != nil {
		return "", fmt.Errorf("error scanning query row full url %w", err)
	}

	if fullURL == "" {
		return fullURL, errNotFound
	}
	return fullURL, nil
}

func (db *DatabaseStorage) CheckShort(ctx context.Context, shortURL string) bool {
	if _, err := db.GetFullURL(ctx, shortURL); err != nil {
		return false
	}

	return true
}

func (db *DatabaseStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("error ping context %w", err)
	}

	return nil
}
