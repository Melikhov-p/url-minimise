package storage

import (
	"context"
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAddURL(t *testing.T) {
	storage := NewMemoryStorage()

	newURL := &models.StorageURL{
		ShortURL:    "short",
		OriginalURL: "original",
		UserID:      1,
		DeletedFlag: false,
	}

	shortURL, err := storage.AddURL(context.Background(), newURL)
	assert.NoError(t, err)
	assert.Equal(t, "short", shortURL)

	// Проверка, что URL добавлен в хранилище
	storedURL, err := storage.GetURL(context.Background(), "short")
	assert.NoError(t, err)
	assert.Equal(t, newURL, storedURL)
}

func TestAddURLs(t *testing.T) {
	storage := NewMemoryStorage()

	newURLs := []*models.StorageURL{
		{ShortURL: "short1", OriginalURL: "original1", UserID: 1, DeletedFlag: false},
		{ShortURL: "short2", OriginalURL: "original2", UserID: 1, DeletedFlag: false},
	}

	err := storage.AddURLs(context.Background(), newURLs)
	assert.NoError(t, err)

	// Проверка, что все URL добавлены в хранилище
	for _, url := range newURLs {
		storedURL, err := storage.GetURL(context.Background(), url.ShortURL)
		assert.NoError(t, err)
		assert.Equal(t, url, storedURL)
	}
}

func TestAddDeleteTask(t *testing.T) {
	storage := NewMemoryStorage()

	shortURLs := []string{"short1", "short2"}
	userID := 1

	err := storage.AddDeleteTask(shortURLs, userID)
	assert.NoError(t, err)

	// Проверка, что задачи на удаление добавлены
	tasks, err := storage.GetDeleteTasksWStatus(context.Background(), models.Registered)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestGetDeleteTasksWStatus(t *testing.T) {
	storage := NewMemoryStorage()

	_ = storage.AddDeleteTask([]string{"short1"}, 1)
	_ = storage.AddDeleteTask([]string{"short2"}, 1)

	tasks, err := storage.GetDeleteTasksWStatus(context.Background(), models.Registered)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestMarkAsDeletedURL(t *testing.T) {
	storage := NewMemoryStorage()

	newURL := &models.StorageURL{
		ShortURL:    "short",
		OriginalURL: "original",
		UserID:      1,
		DeletedFlag: false,
	}

	_, err := storage.AddURL(context.Background(), newURL)
	assert.NoError(t, err)

	task := &models.DelTask{
		URL:    "short",
		UserID: 1,
		Status: models.Registered,
	}

	err = storage.MarkAsDeletedURL(context.Background(), []*models.DelTask{task})
	assert.NoError(t, err)

	// Проверка, что URL отмечен как удаленный
	storedURL, err := storage.GetURL(context.Background(), "short")
	assert.NoError(t, err)
	assert.True(t, storedURL.DeletedFlag)
}

func TestUpdateTasksStatus(t *testing.T) {
	storage := NewMemoryStorage()

	_ = storage.AddDeleteTask([]string{"short1"}, 1)
	_ = storage.AddDeleteTask([]string{"short2"}, 1)

	tasks, err := storage.GetDeleteTasksWStatus(context.Background(), models.Registered)
	assert.NoError(t, err)

	newStatus := models.Done
	err = storage.UpdateTasksStatus(context.Background(), tasks, newStatus)
	assert.NoError(t, err)

	// Проверка, что статус задач обновлен
	updatedTasks, err := storage.GetDeleteTasksWStatus(context.Background(), newStatus)
	assert.NoError(t, err)
	assert.Len(t, updatedTasks, 2)
}

func TestGetShortURL(t *testing.T) {
	storage := NewMemoryStorage()

	newURL := &models.StorageURL{
		ShortURL:    "short",
		OriginalURL: "original",
		UserID:      1,
		DeletedFlag: false,
	}

	_, err := storage.AddURL(context.Background(), newURL)
	assert.NoError(t, err)

	shortURL, err := storage.GetShortURL(context.Background(), nil, "original")
	assert.NoError(t, err)
	assert.Equal(t, "short", shortURL)
}

func TestGetURL(t *testing.T) {
	storage := NewMemoryStorage()

	newURL := &models.StorageURL{
		ShortURL:    "short",
		OriginalURL: "original",
		UserID:      1,
		DeletedFlag: false,
	}

	_, err := storage.AddURL(context.Background(), newURL)
	assert.NoError(t, err)

	storedURL, err := storage.GetURL(context.Background(), "short")
	assert.NoError(t, err)
	assert.Equal(t, newURL, storedURL)

	_, err = storage.GetURL(context.Background(), "unStored")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestCheckShort(t *testing.T) {
	storage := NewMemoryStorage()

	newURL := &models.StorageURL{
		ShortURL:    "short",
		OriginalURL: "original",
		UserID:      1,
		DeletedFlag: false,
	}

	_, err := storage.AddURL(context.Background(), newURL)
	assert.NoError(t, err)

	exists := storage.CheckShort(context.Background(), "short")
	assert.True(t, exists)
}

func TestAddUser(t *testing.T) {
	storage := NewMemoryStorage()

	user, err := storage.AddUser(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)

	// Проверка, что пользователь добавлен в хранилище
	storedUser := storage.users[1]
	assert.Equal(t, user, storedUser)
}

func TestMemoryStorage_GetURLsByUserID(t *testing.T) {
	storage := NewMemoryStorage()

	user, err := storage.AddUser(context.Background())
	assert.NoError(t, err)

	_, err = storage.GetURLsByUserID(context.Background(), user.ID)
	assert.NoError(t, err)
}

func TestMemoryStorage_Ping(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.Ping(context.Background())
	assert.NoError(t, err)
}

func TestNewMemoryStorage_checkFull(t *testing.T) {
	storage := NewMemoryStorage()

	_, ok := storage.checkFull(context.Background(), "full")
	assert.False(t, ok)
}

func TestMemoryStorage_Close(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.Close()

	assert.NoError(t, err)
}

func TestMemoryStorage_GetURLsCount(t *testing.T) {
	storage := NewMemoryStorage()

	count, err := storage.GetURLsCount(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMemoryStorage_GetUsersCount(t *testing.T) {
	storage := NewMemoryStorage()

	count, err := storage.GetUsersCount(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
