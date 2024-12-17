package app

import (
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/handlers"
	"github.com/Melikhov-p/url-minimise/internal/middlewares"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func CreateRouter(cfg *config.Config, storage repository.Storage, logger *zap.Logger) chi.Router {
	router := chi.NewRouter()
	middleware := middlewares.Middleware{
		Logger: logger,
	}

	router.Use(
		middleware.WithLogging,
		middleware.GzipMiddleware,
	)

	router.Get("/ping", wrapper(handlers.PingDatabase, cfg, storage, logger))

	router.Post("/", wrapper(handlers.CreateShortURL, cfg, storage, logger))

	router.Get("/{id}", wrapper(handlers.GetFullURL, cfg, storage, logger))

	router.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", wrapper(handlers.APICreateShortURL, cfg, storage, logger))
			r.Post("/batch", wrapper(handlers.APICreateBatchURLs, cfg, storage, logger))
		})
		r.Route("/user", func(r chi.Router) {
			r.Get("/urls", wrapper(handlers.GetUserURLs, cfg, storage, logger))
			r.Delete("/urls", wrapper(handlers.APIMarkAsDeletedURLs, cfg, storage, logger))
		})
	})

	return router
}

func wrapper(
	wrappedFunc func(http.ResponseWriter,
		*http.Request,
		*config.Config,
		repository.Storage,
		*zap.Logger),
	cfg *config.Config,
	storage repository.Storage,
	logger *zap.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wrappedFunc(w, r, cfg, storage, logger)
	}
}
