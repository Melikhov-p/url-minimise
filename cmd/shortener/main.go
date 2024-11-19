package main

import (
	"log"
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/app"
	"github.com/Melikhov-p/url-minimise/internal/config"
	loggerBuilder "github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	cfg.Build()

	logger, err := loggerBuilder.BuildLogger("DEBUG")
	if err != nil {
		log.Fatalf("cannot run logger: %v", err)
	}

	storage, err := models.NewStorage(models.STORAGEFROMFILE, cfg, logger)
	if err != nil {
		logger.Fatal("error building storage from file", zap.Error(err))
	}

	router := app.CreateRouter(cfg, storage, logger)

	logger.Info("Running server on", zap.String("address", cfg.ServerAddr))
	err = http.ListenAndServe(cfg.ServerAddr, router)

	if err != nil {
		logger.Fatal("Fatal error", zap.Error(err))
	}
}
