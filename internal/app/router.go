package app

import (
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/handlers"
	"github.com/Melikhov-p/url-minimise/internal/middlewares"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/go-chi/chi/v5"
)

func CreateRouter(cfg *config.Config, storage *models.Storage) chi.Router {
	router := chi.NewRouter()

	router.Post(
		"/", middlewares.Conveyor(
			func(w http.ResponseWriter, r *http.Request) {
				handlers.CreateShortURL(w, r, cfg, storage)
			},
			middlewares.GzipMiddleware,
			middlewares.WithLogging,
		))
	router.Get("/{id}", middlewares.Conveyor(
		func(w http.ResponseWriter, r *http.Request) {
			handlers.GetFullURL(w, r, storage)
		},
		middlewares.GzipMiddleware,
		middlewares.WithLogging))
	router.Route("/api", func(r chi.Router) {
		r.Post("/shorten", middlewares.Conveyor(
			func(w http.ResponseWriter, r *http.Request) {
				handlers.APICreateShortURL(w, r, cfg, storage)
			},
			middlewares.GzipMiddleware,
			middlewares.WithLogging,
		))
	})

	return router
}
