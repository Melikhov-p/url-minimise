package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"time"

	"github.com/Melikhov-p/url-minimise/internal/app"
	"github.com/Melikhov-p/url-minimise/internal/config"
	loggerBuilder "github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/Melikhov-p/url-minimise/internal/worker"
	"go.uber.org/zap"
)

// BuildVersion = определяет версию приложения.
// BuildDate = определяет дату сборки.
// BuildCommit = определяет коммит сборки.
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

const delWorkerPingInterval = 5 * time.Second

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
	logger, err := loggerBuilder.BuildLogger("DEBUG")
	if err != nil {
		log.Fatalf("cannot run logger: %v", err)
	}

	cfg := config.NewConfig(logger, false)

	store, err := repository.NewStorage(cfg, logger)
	if err != nil {
		logger.Fatal("error getting new storage", zap.Error(err))
	}

	delWorker := worker.NewDelWorker(delWorkerPingInterval, logger, store)
	go delWorker.LookUp()

	router := app.CreateRouter(cfg, store, logger)

	logger.Info("Running server on", zap.String("address", cfg.ServerAddr))
	err = http.ListenAndServe(cfg.ServerAddr, router)

	if err != nil {
		logger.Fatal("Fatal error", zap.Error(err))
	}
}
