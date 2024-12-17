package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/auth"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"go.uber.org/zap"
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
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	shortURL, err := db.GetShortURL(ctx, tx, newURL.OriginalURL)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("error getting short URL for original %w", err)
	}

	if shortURL != "" {
		return shortURL, ErrOriginalURLExist
	}

	preparedInsert, err := tx.PrepareContext(ctx, `
		INSERT INTO URL (short_url, original_url, user_id, is_deleted)
        VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		return "", fmt.Errorf("error creating prepared insert query %w", err)
	}

	defer func() {
		_ = preparedInsert.Close()
	}()

	_, err = preparedInsert.ExecContext(ctx, newURL.ShortURL, newURL.OriginalURL, newURL.UserID, newURL.DeletedFlag)
	if err != nil {
		return "", fmt.Errorf("error exec context from database in addurl %w", err)
	}

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

	placeholders := make([]string, len(newURLs))
	values := make([]interface{}, 0, len(newURLs)*4)
	for i, url := range newURLs {
		placeholders[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4)
		values = append(values, url.ShortURL, url.OriginalURL, url.UserID, url.DeletedFlag)
	}

	preparedInsert, err := tx.PrepareContext(ctx, fmt.Sprintf(`
		INSERT INTO url (short_url, original_url, user_id, is_deleted) VALUES %s
	`, strings.Join(placeholders, ", ")))
	if err != nil {
		return fmt.Errorf("error prepare insert query for multi urls %w", err)
	}
	defer func() {
		_ = preparedInsert.Close()
	}()

	_, err = preparedInsert.ExecContext(ctx, values...)
	if err != nil {
		return fmt.Errorf("error executing context for prepared insert %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction %w", err)
	}
	return nil
}

func (db *DatabaseStorage) MarkAsDeletedURL(ctx context.Context,
	urls []string,
	userID int,
	logger *zap.Logger) error {
	placeholders := make([]string, len(urls))
	for i := range urls {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	query := fmt.Sprintf("UPDATE url SET is_deleted=true WHERE short_url IN (%s) AND user_id = $%d", strings.Join(placeholders, ", "), len(urls)+1)

	// Передаем каждый URL как отдельный параметр
	args := make([]interface{}, len(urls)+1)
	for i, url := range urls {
		args[i] = url
	}
	args[len(urls)] = userID

	if _, err := db.DB.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("error executing context for update query: %w", err)
	} else {
		logger.Debug("updated record in database for URLs", zap.Strings("URLs", urls))
		return nil
	}
}

func (db *DatabaseStorage) GetURL(ctx context.Context, shortURL string) (*models.StorageURL, error) {
	query := `
		SELECT original_url, user_id, uuid, is_deleted  FROM url WHERE short_url = $1
	`

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var u models.StorageURL
	u.ShortURL = shortURL
	if err := db.DB.QueryRowContext(ctx, query, shortURL).
		Scan(&u.OriginalURL, &u.UserID, &u.UUID, &u.DeletedFlag); err != nil {
		return nil, fmt.Errorf("error scanning query row full url %w", err)
	}

	return &u, nil
}

func (db *DatabaseStorage) GetShortURL(ctx context.Context, tx *sql.Tx, fullURL string) (string, error) {
	preparedSelect, err := tx.PrepareContext(ctx, `SELECT short_url FROM url WHERE original_url = $1`)
	if err != nil {
		return "", fmt.Errorf("error prepare select query %w", err)
	}
	defer func() {
		_ = preparedSelect.Close()
	}()

	var shortURL string
	rows := preparedSelect.QueryRowContext(ctx, fullURL)
	if err = rows.Scan(&shortURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return shortURL, ErrNotFound
		}
		return shortURL, fmt.Errorf("error scanning rows from database %w", err)
	}
	return shortURL, nil
}

func (db *DatabaseStorage) CheckShort(ctx context.Context, shortURL string) bool {
	if _, err := db.GetURL(ctx, shortURL); err != nil {
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

func (db *DatabaseStorage) AddUser(ctx context.Context) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `INSERT INTO "user" DEFAULT VALUES RETURNING id`

	var id int
	row := tx.QueryRowContext(ctx, query)
	if err = row.Scan(&id); err != nil {
		return nil, fmt.Errorf("error insert new user in db %w", err)
	}

	user := &models.User{
		ID:   id,
		URLs: make([]*models.StorageURL, 0),
		Service: &models.UserService{
			IsAuthenticated: false,
			Token:           "",
		},
	}

	_ = tx.Commit()
	return user, nil
}

func (db *DatabaseStorage) GetURLsByUserID(ctx context.Context, userID int) ([]*models.StorageURL, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
				SELECT short_url, original_url, uuid
				FROM url WHERE user_id = $1;`

	rows, err := db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting url rows from db %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	urls := make([]*models.StorageURL, 0)
	for rows.Next() {
		var url models.StorageURL
		if err = rows.Scan(&url.ShortURL, &url.OriginalURL, &url.UUID); err != nil {
			return nil, fmt.Errorf("error scanning url from db response %w", err)
		}
		urls = append(urls, &url)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating for rows of urls %w", err)
	}

	return urls, nil
}

func (db *DatabaseStorage) GetSecretKey(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tx, err := db.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("error starting transaction for secret kay %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `SELECT key FROM secret_key`

	row := tx.QueryRowContext(ctx, query)

	var key string
	if err = row.Scan(&key); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("error scanning secret key row from db %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		key, err = auth.GenerateAuthKey()
		if err != nil {
			return "", fmt.Errorf("error generating new secret key %w", err)
		}

		insertQuery := `INSERT INTO secret_key (key) VALUES ($1)`
		_, err = tx.ExecContext(ctx, insertQuery, key)
		if err != nil {
			return "", fmt.Errorf("error exec context %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("error commiting %w", err)
	}
	return key, nil
}
