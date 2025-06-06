package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"os/signal"
	"syscall"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/app"
	"github.com/Melikhov-p/url-minimise/internal/config"
	grpc2 "github.com/Melikhov-p/url-minimise/internal/grpc"
	loggerBuilder "github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/middlewares"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/Melikhov-p/url-minimise/internal/worker"
	"github.com/Melikhov-p/url-minimise/protos/gen/proto"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// BuildVersion = определяет версию приложения.
// BuildDate = определяет дату сборки.
// BuildCommit = определяет коммит сборки.
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

const delWorkerPingInterval = 4 * time.Second
const (
	timeoutServerShutdown = time.Second * 5
	timeoutShutdown       = time.Second * 10
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

// Run start all
func Run() (err error) {
	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancelCtx()

	eg, ctx := errgroup.WithContext(rootCtx)
	// нештатное завершение программы по таймауту
	// происходит, если после завершения контекста
	// приложение не смогло завершиться за отведенный промежуток времени
	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	logger, err := loggerBuilder.BuildLogger("DEBUG")
	if err != nil {
		return fmt.Errorf("error building logger %w", err)
	}

	cfg := config.NewConfig(logger, false)

	store, err := repository.NewStorage(cfg, logger)
	if err != nil {
		logger.Fatal("error getting new storage", zap.Error(err))
	}
	eg.Go(func() error {
		defer logger.Debug("DB closed")

		<-ctx.Done()

		err = store.Close()

		logger.Error("error closing storage", zap.Error(err))
		return nil
	})

	router := app.CreateRouter(cfg, store, logger)

	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	eg.Go(func() error {
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
			if err = server.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("error listen and server: %w", err)
			}
			return nil
		}

		logger.Info("Running server on", zap.String("address", cfg.ServerAddr), zap.Bool("TLS", false))
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("error listen and server: %w", err)
		}
		return nil
	})

	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		logger.Error("error running gRPC listnere", zap.Error(err))
		<-ctx.Done()
	}
	serverRPC := grpc.NewServer(
		grpc.UnaryInterceptor(
			middlewares.NewUnaryInterceptor(
				logger,
				cfg,
				store,
			).UnaryAuthInterceptor,
		),
	)
	proto.RegisterShortenerServer(serverRPC, grpc2.NewShortenerService(logger, cfg, store))

	eg.Go(func() error {
		logger.Debug("gRPC server is running on", zap.String("port", ":3200"))

		if err := serverRPC.Serve(listen); err != nil {
			return fmt.Errorf("error in gRPC server %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer logger.Debug("gRPC server has been stopped.")
		<-ctx.Done()

		serverRPC.GracefulStop()

		return nil
	})

	eg.Go(func() error {
		defer logger.Debug("server has been shutdown")
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err = server.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Printf("an error occurred during server shutdown: %v", err)
		}
		return nil
	})

	delWorker := worker.NewDelWorker(delWorkerPingInterval, logger, store)

	eg.Go(func() error {
		delWorker.LookUp()
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()

		delWorker.Stop()
		return nil
	})

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("errgroup error: %w", err)
	}

	return nil
}
