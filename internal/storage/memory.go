package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

// MemoryStorage хранилище в памяти.
type MemoryStorage struct {
	urls        map[string]*models.StorageURL
	users       map[int]*models.User
	deleteTasks map[string]*models.DelTask // [shortURL]*models.DelTask
	lastUserID  int
}

// NewMemoryStorage создать новое хранилище в памяти.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls:        map[string]*models.StorageURL{},
		users:       map[int]*models.User{},
		deleteTasks: map[string]*models.DelTask{},
		lastUserID:  0,
	}
}

// AddURL добавить адрес.
func (s *MemoryStorage) AddURL(ctx context.Context, newURL *models.StorageURL) (string, error) {
	if short, ok := s.checkFull(ctx, newURL.OriginalURL); ok {
		return short, ErrOriginalURLExist
	}
	s.urls[newURL.ShortURL] = newURL
	return newURL.ShortURL, nil
}

// AddURLs добавить несколько адресов.
func (s *MemoryStorage) AddURLs(_ context.Context, newURLs []*models.StorageURL) error {
	for _, url := range newURLs {
		s.urls[url.ShortURL] = url
	}

	return nil
}

// AddDeleteTask добавить задачу на удаление.
func (s *MemoryStorage) AddDeleteTask(shortURL []string, userID int) error {
	for _, url := range shortURL {
		s.deleteTasks[url] = &models.DelTask{
			URL:    url,
			UserID: userID,
			Status: models.Registered,
		}
	}

	return nil
}

// GetDeleteTasksWStatus получить статус задачи на удаление.
func (s *MemoryStorage) GetDeleteTasksWStatus(
	_ context.Context,
	status models.DelTaskStatus,
) ([]*models.DelTask, error) {
	outTasks := make([]*models.DelTask, 0)
	for _, task := range s.deleteTasks {
		if task.Status == status {
			outTasks = append(outTasks, task)
		}
	}

	return outTasks, nil
}

// MarkAsDeletedURL отметить адрес на удаление.
func (s *MemoryStorage) MarkAsDeletedURL(_ context.Context, tasks []*models.DelTask) error {
	for _, task := range tasks {
		if s.urls[task.URL] == nil {
			return ErrNotFound
		}
		if s.urls[task.URL].UserID != task.UserID {
			return fmt.Errorf("not owner of %s", task.URL)
		}

		s.urls[task.URL].DeletedFlag = true
	}

	return nil
}

// UpdateTasksStatus обновить статус задачи на удаление.
func (s *MemoryStorage) UpdateTasksStatus(
	_ context.Context,
	tasks []*models.DelTask,
	newStatus models.DelTaskStatus,
) error {
	for _, task := range tasks {
		if _, ok := s.deleteTasks[task.URL]; ok {
			s.deleteTasks[task.URL].Status = newStatus
		}
	}

	return nil
}

// GetShortURL получить короткий адрес.
func (s *MemoryStorage) GetShortURL(_ context.Context, _ *sql.Tx, fullURL string) (string, error) {
	var short string

	for _, url := range s.urls {
		if url.OriginalURL == fullURL {
			short = url.ShortURL
			return short, nil
		}
	}

	return "", fmt.Errorf("can not found short url for original %w", ErrNotFound)
}

// GetURL получить полный адрес.
func (s *MemoryStorage) GetURL(_ context.Context, shortURL string) (*models.StorageURL, error) {
	searchedElem := s.urls[shortURL]
	if searchedElem != nil {
		return searchedElem, nil
	}
	return nil, fmt.Errorf("can not found original url for short %w", ErrNotFound)
}

// CheckShort проверить короткий адрес.
func (s *MemoryStorage) CheckShort(_ context.Context, short string) bool { return s.urls[short] != nil }

// Если оригинальный URL есть в базе - true.
func (s *MemoryStorage) checkFull(_ context.Context, fullURL string) (string, bool) {
	for _, url := range s.urls {
		if fullURL == url.OriginalURL {
			return url.ShortURL, true
		}
	}

	return "", false
}

// Ping пинг
func (s *MemoryStorage) Ping(_ context.Context) error { return nil }

// AddUser добавить пользователя
func (s *MemoryStorage) AddUser(_ context.Context) (*models.User, error) {
	s.lastUserID++
	s.users[s.lastUserID] = &models.User{
		ID:   s.lastUserID,
		URLs: make([]*models.StorageURL, 0),
		Service: &models.UserService{
			IsAuthenticated: false,
			Token:           "",
		},
	}

	return s.users[s.lastUserID], nil
}

// GetURLsByUserID получить адреса пользователя.
func (s *MemoryStorage) GetURLsByUserID(_ context.Context, userID int) ([]*models.StorageURL, error) {
	user := s.users[userID]
	if user != nil {
		urls := user.URLs
		if urls != nil {
			return urls, nil
		}
	}
	return []*models.StorageURL{}, nil
}
