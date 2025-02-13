package service

import (
	"context"
	"testing"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/repository"
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
