package main

import (
	"log"
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/app"
	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	cfg.Build()

	if err := logger.Initialize("DEBUG"); err != nil {
		log.Fatal("cannot run logger " + err.Error())
	}

	router := app.CreateRouter(cfg)

	logger.Log.Info("Running server on", zap.String("address", cfg.ServerAddr))
	err := http.ListenAndServe(cfg.ServerAddr, router)

	if err != nil {
		logger.Log.Fatal("Fatal error", zap.Error(err))
	}
}
