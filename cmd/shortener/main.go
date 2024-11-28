package main

import (
	"log"
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/app"
	"github.com/Melikhov-p/url-minimise/internal/config"
	loggerBuilder "github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig() // Возвращает конфиг со значениями по-умолчанию
	cfg.Build()               // Забирает значения флагов и переменных окружения

	logger, err := loggerBuilder.BuildLogger("DEBUG")
	if err != nil {
		log.Fatalf("cannot run logger: %v", err)
	}

	store, err := repository.NewStorage(cfg)
	if err != nil {
		logger.Fatal("error getting new storage", zap.Error(err))
	}

	router := app.CreateRouter(cfg, store, logger)

	logger.Info("Running server on", zap.String("address", cfg.ServerAddr))
	err = http.ListenAndServe(cfg.ServerAddr, router)

	if err != nil {
		logger.Fatal("Fatal error", zap.Error(err))
	}
}
