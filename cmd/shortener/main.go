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
	"golang.org/x/crypto/acme/autocert"
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

	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router}

	if cfg.TLS {
		manager := autocert.Manager{
			// директория для хранения сертификатов
			Cache: autocert.DirCache("cache-dir"),
			// функция, принимающая Terms of Service издателя сертификатов
			Prompt: autocert.AcceptTOS,
			// перечень доменов, для которых будут поддерживаться сертификаты
			HostPolicy: autocert.HostWhitelist("mysite.ru", "www.mysite.ru"),
		}
		server.TLSConfig = manager.TLSConfig()
		logger.Info("Running server on", zap.String("address", cfg.ServerAddr), zap.Bool("TLS", true))
		err = server.ListenAndServeTLS("", "")
	} else {
		logger.Info("Running server on", zap.String("address", cfg.ServerAddr), zap.Bool("TLS", false))
		err = server.ListenAndServe()
	}

	if err != nil {
		logger.Fatal("Fatal error", zap.Error(err))
	}
}
