package repository

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
	storageConfig "github.com/Melikhov-p/url-minimise/internal/repository/config"
	fileConfig "github.com/Melikhov-p/url-minimise/internal/repository/file/config"
	memoryConfig "github.com/Melikhov-p/url-minimise/internal/repository/memory/config"
	storage2 "github.com/Melikhov-p/url-minimise/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewStorage(t *testing.T) {
	logger := zap.NewNop()

	t.Run("MemoryStorage", func(t *testing.T) {
		cfg := &config.Config{
			StorageMode: storage2.BaseStorage,
		}

		storage, err := NewStorage(cfg, logger)
		assert.NoError(t, err)
		assert.NotNil(t, storage)
		assert.NotEmpty(t, cfg.SecretKey)
	})

	t.Run("FileStorage", func(t *testing.T) {
		cfg := &config.Config{
			StorageMode: storage2.StorageFromFile,
			Storage: storageConfig.Config{
				InMemory: &memoryConfig.Config{},
				FileStorage: &fileConfig.Config{
					FilePath: "storage.txt",
				},
			},
		}

		// Создаем временный файл для теста
		file, err := os.CreateTemp("", "testfile.txt")
		assert.NoError(t, err)
		defer func() {
			_ = os.Remove(file.Name())
		}()

		cfg.Storage.FileStorage.FilePath = file.Name()

		// Записываем данные в файл
		element := models.StorageURL{ShortURL: "test"}
		data, _ := json.Marshal(element)
		_, err = file.Write(data)
		assert.NoError(t, err)
		_, err = file.Seek(0, 0)
		assert.NoError(t, err)

		storage, err := NewStorage(cfg, logger)
		assert.NoError(t, err)
		assert.NotNil(t, storage)
		assert.NotEmpty(t, cfg.SecretKey)
	})
}

func TestNewStorageURL(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		StorageMode: storage2.BaseStorage,
	}
	storage, err := NewStorage(cfg, logger)
	assert.NoError(t, err)
	su, err := NewStorageURL(context.Background(), "full", storage, cfg, 1)
	assert.NoError(t, err)
	assert.Equal(t, su.UserID, 1)
}

func TestNewStorageMultiURL(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		StorageMode: storage2.BaseStorage,
	}
	storage, err := NewStorage(cfg, logger)
	assert.NoError(t, err)
	su, err := NewStorageMultiURL(context.Background(), []string{"full", "full2", "full3"}, storage, cfg, 1)
	assert.NoError(t, err)
	assert.Equal(t, su[0].UserID, 1)
}

func TestRandomString(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		StorageMode: storage2.BaseStorage,
	}
	storage, err := NewStorage(cfg, logger)
	assert.NoError(t, err)
	str, err := randomString(context.Background(), 10, storage)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(str))
}
