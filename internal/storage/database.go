package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

type DatabaseStorage struct {
	DB *sql.DB
}

const dbTimeout = 15 * time.Second

func (db *DatabaseStorage) AddURL(ctx context.Context, newURL *models.StorageURL) (string, error) {
	// Add new url in storage, return short url and error.
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tx, err := db.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("error starting transaction %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	preparedInsert, err := tx.PrepareContext(ctx, `
		INSERT INTO URL (short_url, original_url)
        VALUES ($1, $2)
	`)
	if err != nil {
		return "", fmt.Errorf("error creating prepared insert query %w", err)
	}

	defer func() {
		_ = preparedInsert.Close()
	}()

	_, err = preparedInsert.ExecContext(ctx, newURL.ShortURL, newURL.OriginalURL)
	if err != nil {
		if strings.Contains(err.Error(), UniqueViolationCode) {
			if strings.Contains(err.Error(), errOriginalURLExist.Field) {
				shortURL, err := db.GetShortURL(ctx, tx, newURL.OriginalURL)
				if err != nil {
					return "", fmt.Errorf("error getting existing short url %w", err)
				}
				return shortURL, fmt.Errorf("%w : %s", ErrOriginalURLExist, newURL.OriginalURL)
			}
		}
		return "", fmt.Errorf("error exec context from database in addurl %w", err)
	}

	_ = tx.Commit()
	return newURL.ShortURL, nil
}

func (db *DatabaseStorage) AddURLs(ctx context.Context, newURLs []*models.StorageURL) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("error begin transaction %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	preparedQuery, err := tx.PrepareContext(ctx, `
		INSERT INTO url (short_url, original_url) VALUES ($1, $2)
	`)
	if err != nil {
		if strings.Contains(err.Error(), UniqueViolationCode) {
			if strings.Contains(err.Error(), errOriginalURLExist.Field) {
				return ErrOriginalURLExist
			}
		}
		return fmt.Errorf("error creating prepared query in addURLs %w", err)
	}
	defer func() {
		_ = preparedQuery.Close()
	}()

	for _, url := range newURLs {
		_, err = preparedQuery.ExecContext(ctx, url.ShortURL, url.OriginalURL)
		if err != nil {
			return fmt.Errorf("error exec context %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction %w", err)
	}
	return nil
}

func (db *DatabaseStorage) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	query := `
		SELECT original_url FROM url WHERE short_url = $1
	`

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var fullURL string
	if err := db.DB.QueryRowContext(ctx, query, shortURL).Scan(&fullURL); err != nil {
		return "", fmt.Errorf("error scanning query row full url %w", err)
	}

	if fullURL == "" {
		return fullURL, ErrNotFound
	}
	return fullURL, nil
}

func (db *DatabaseStorage) GetShortURL(ctx context.Context, tx *sql.Tx, fullURL string) (string, error) {
	preparedSelect, err := tx.PrepareContext(ctx, `SELECT short_url FROM url WHERE original_url = $1`)
	if err != nil {
		return "", err
	}

	var shortURL string
	if err := preparedSelect.QueryRowContext(ctx, fullURL).Scan(&shortURL); err != nil {
		return "", fmt.Errorf("error scanning query row full url %w", err)
	}

	if shortURL == "" {
		return shortURL, ErrNotFound
	}
	return shortURL, nil
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
