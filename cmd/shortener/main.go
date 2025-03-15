package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"os"
	"os/signal"
	"syscall"
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

	router := app.CreateRouter(cfg, store, logger)

	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router}

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

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
		go func() {
			if err = server.ListenAndServeTLS("", ""); err != nil &&
				errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("fatal error in listen and serve", zap.Error(err))
			}
		}()
	} else {
		logger.Info("Running server on", zap.String("address", cfg.ServerAddr), zap.Bool("TLS", false))
		go func() {
			if err = server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("fatal error in listen and serve", zap.Error(err))
			}
		}()
	}

	delWorker := worker.NewDelWorker(delWorkerPingInterval, logger, store)
	go delWorker.LookUp()

	sig := <-stopCh
	logger.Debug("stop signal received", zap.Any("signal", sig))

	delWorker.Stop()

	err = server.Shutdown(context.Background())
	if err != nil {
		logger.Fatal("error shutdown server", zap.Error(err))
	}

	logger.Debug("graceful shutdown complete")
}
