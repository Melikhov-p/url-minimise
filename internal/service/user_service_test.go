package service

import (
	"context"
	"testing"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/stretchr/testify/assert"
)

func BenchmarkServices(b *testing.B) {
	log, err := logger.BuildLogger("ERROR")
	if err != nil {
		panic(err.Error())
	}
	store, err := repository.NewStorage(config.NewConfig(log, true), log)
	if err != nil {
		panic(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cfg := config.NewConfig(log, true)
	userID := 999
	var userToken string
	originalURL := "original.url/1"
	b.ResetTimer()

	b.Run("BuildUserToken", func(b *testing.B) {
		userToken, _ = BuildUserToken(userID, cfg)
	})

	b.Run("Auth by token", func(b *testing.B) {
		_, _ = AuthUserByToken(userToken, store, log, cfg)
	})

	b.Run("ADD URL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = AddURL(ctx, store, log, originalURL, cfg, userID)
		}
	})
}

func TestAddNewUser(t *testing.T) {
	log, err := logger.BuildLogger("ERROR")
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)
	store, err := repository.NewStorage(cfg, log)
	assert.NoError(t, err)
	_, err = AddNewUser(context.Background(), store, cfg)
	assert.NoError(t, err)
}

func TestBuildUserToken(t *testing.T) {
	log, err := logger.BuildLogger("ERROR")
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)
	_, err = BuildUserToken(1, cfg)
	assert.NoError(t, err)
}

func TestAuthUserByToken(t *testing.T) {
	log, err := logger.BuildLogger("ERROR")
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)
	store, err := repository.NewStorage(cfg, log)
	assert.NoError(t, err)

	token, err := BuildUserToken(1, cfg)
	assert.NoError(t, err)

	user, err := AuthUserByToken(token, store, log, cfg)
	assert.NoError(t, err)
	assert.Equal(t, token, user.Service.Token)
}
