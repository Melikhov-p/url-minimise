package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSetInMemory(t *testing.T) {
	// Создаем временный файл для теста
	file, err := os.CreateTemp("", "testfile.txt")
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(file.Name())
	}()

	storage := &FileStorage{
		MemoryStorage: *NewMemoryStorage(),
		File:          file,
		Encoder:       json.NewEncoder(file),
		Scanner:       bufio.NewScanner(file),
	}

	newURL := &models.StorageURL{
		ShortURL:    "short",
		OriginalURL: "original",
		UserID:      1,
		DeletedFlag: false,
	}

	storage.SetInMemory("short", newURL)

	// Проверка, что URL добавлен в in-memory хранилище
	storedURL := storage.urls["short"]
	assert.Equal(t, newURL, storedURL)
}

func TestSave(t *testing.T) {
	// Создаем временный файл для теста
	file, err := os.CreateTemp("", "testfile.txt")
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(file.Name())
	}()

	storage := &FileStorage{
		MemoryStorage: *NewMemoryStorage(),
		File:          file,
		Encoder:       json.NewEncoder(file),
		Scanner:       bufio.NewScanner(file),
	}

	newURL := &models.StorageURL{
		ShortURL:    "short",
		OriginalURL: "original",
		UserID:      1,
		DeletedFlag: false,
	}

	err = storage.Save(newURL)
	assert.NoError(t, err)

	// Проверка, что данные записаны в файл
	_, err = file.Seek(0, 0)
	assert.NoError(t, err)

	var savedURL models.StorageURL
	err = json.NewDecoder(file).Decode(&savedURL)
	assert.NoError(t, err)
	assert.Equal(t, *newURL, savedURL)
}
