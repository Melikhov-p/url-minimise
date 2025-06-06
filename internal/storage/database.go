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
)

// DatabaseStorage хранилище в базе данных.
type DatabaseStorage struct {
	DB *sql.DB
}

const dbTimeout = 15 * time.Second

// Close connection.
func (db *DatabaseStorage) Close() error {
	err := db.DB.Close()
	if err != nil {
		return fmt.Errorf("error closing db %w", err)
	}

	return nil
}

// AddURL добавить URL.
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

// AddURLs добавить несколько URL.
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

// AddDeleteTask добавит задачу на удаление.
func (db *DatabaseStorage) AddDeleteTask(
	shortURL []string,
	userID int,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `INSERT INTO delete_task (short_url, user_id, status) VALUES ($1, $2, $3)`
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction for create delete task %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error prepare context for create del task %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, url := range shortURL {
		_, err = stmt.ExecContext(ctx, url, userID, models.Registered)
		if err != nil {
			return fmt.Errorf("error exec context for create del task %w", err)
		}
	}

	_ = tx.Commit()
	return nil
}

// GetDeleteTasksWStatus получить статус задачи на удаление.
func (db *DatabaseStorage) GetDeleteTasksWStatus(
	ctx context.Context,
	status models.DelTaskStatus,
) ([]*models.DelTask, error) {
	query := `SELECT short_url, user_id FROM delete_task WHERE status=$1`
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	rows, err := db.DB.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("error exec context for delete tasks %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	outTasks := make([]*models.DelTask, 0)
	for rows.Next() {
		var task models.DelTask
		if err = rows.Scan(&task.URL, &task.UserID); err != nil {
			return nil, fmt.Errorf("error scanning rows for tasks %w", err)
		}
		task.Status = status
		outTasks = append(outTasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err() return error %w", err)
	}

	return outTasks, nil
}

// MarkAsDeletedURL отметить адрес на удаление
func (db *DatabaseStorage) MarkAsDeletedURL(ctx context.Context, tasks []*models.DelTask) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	// Прохожу по таскам в цикле, а не отправляю батчем, потому что в теории у тасок может быть разный user_id
	query := `UPDATE url SET is_deleted=true WHERE short_url=$1 AND user_id=$2`

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction for delete %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error prepare context for update query %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, task := range tasks {
		_, err = stmt.ExecContext(ctx, task.URL, task.UserID)
		if err != nil {
			return fmt.Errorf("error updating record for delete %w", err)
		}
	}

	_ = tx.Commit()
	return nil
}

// UpdateTasksStatus обновить статус задачи на удаление.
func (db *DatabaseStorage) UpdateTasksStatus(
	ctx context.Context,
	tasks []*models.DelTask,
	newStatus models.DelTaskStatus,
) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `UPDATE delete_task SET status=$1 WHERE short_url=$2`

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction for update status %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error prepare context for update task status %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, task := range tasks {
		_, err = stmt.ExecContext(ctx, newStatus, task.URL)
		if err != nil {
			return fmt.Errorf("error updating delete task status %w", err)
		}
	}

	_ = tx.Commit()
	return nil
}

// GetURL получить полный адрес
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

// GetShortURL получить короткий адрес
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

// CheckShort проверить наличие короткого адреса.
func (db *DatabaseStorage) CheckShort(ctx context.Context, shortURL string) bool {
	if _, err := db.GetURL(ctx, shortURL); err != nil {
		return false
	}

	return true
}

// Ping пингануть
func (db *DatabaseStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("error ping context %w", err)
	}

	return nil
}

// AddUser добавить пользователя
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

// GetURLsByUserID получить адреса пользователя
func (db *DatabaseStorage) GetURLsByUserID(ctx context.Context, userID int) ([]*models.StorageURL, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
                                SELECT short_url, original_url, uuid
                                FROM url WHERE user_id = $1;`

	rows, err := db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return []*models.StorageURL{}, fmt.Errorf("error getting url rows from db %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	urls := make([]*models.StorageURL, 0)
	for rows.Next() {
		var url models.StorageURL
		if err = rows.Scan(&url.ShortURL, &url.OriginalURL, &url.UUID); err != nil {
			return []*models.StorageURL{}, fmt.Errorf("error scanning url from db response %w", err)
		}
		urls = append(urls, &url)
	}

	if err = rows.Err(); err != nil {
		return []*models.StorageURL{}, fmt.Errorf("error iterating for rows of urls %w", err)
	}

	return urls, nil
}

// GetSecretKey получить секретный ключ
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

// GetURLsCount получить количество URL.
func (db *DatabaseStorage) GetURLsCount(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var (
		count int
		err   error
	)

	row := db.DB.QueryRowContext(ctx, `SELECT COUNT(original_url) from url`)

	if err = row.Scan(&count); err != nil {
		return -1, fmt.Errorf("error scanning URLs count %w", err)
	}

	return count, nil
}

// GetUsersCount получить количество пользователей.
func (db *DatabaseStorage) GetUsersCount(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var (
		count int
		err   error
	)

	row := db.DB.QueryRowContext(ctx, `SELECT COUNT(id) from "user"`)

	if err = row.Scan(&count); err != nil {
		return -1, fmt.Errorf("error scanning user's count %w", err)
	}

	return count, nil
}
