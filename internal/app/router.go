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

	createURLWrapper := func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateShortURL(w, r, cfg, storage, logger)
	}
	getURLWrapper := func(w http.ResponseWriter, r *http.Request) {
		handlers.GetFullURL(w, r, storage, logger)
	}
	createURLAPIWrapper := func(w http.ResponseWriter, r *http.Request) {
		handlers.APICreateShortURL(w, r, cfg, storage, logger)
	}

	router.Post("/", createURLWrapper)

	router.Get("/{id}", getURLWrapper)

	router.Route("/api", func(r chi.Router) {
		r.Post("/shorten", createURLAPIWrapper)
	})

	return router
}
