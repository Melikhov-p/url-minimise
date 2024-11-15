package main

import (
	"log"
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/app"
	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	cfg.Build()

	if err := logger.Initialize("DEBUG"); err != nil {
		log.Fatal("cannot run logger " + err.Error())
	}

	storage, err := models.NewStorageFromFile(cfg)
	if err != nil {
		logger.Log.Fatal("error building storage from file", zap.Error(err))
	}

	router := app.CreateRouter(cfg, storage)

	logger.Log.Info("Running server on", zap.String("address", cfg.ServerAddr))
	err = http.ListenAndServe(cfg.ServerAddr, router)

	if err != nil {
		logger.Log.Fatal("Fatal error", zap.Error(err))
	}
}
