package app

import (
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/handlers"
	"github.com/Melikhov-p/url-minimise/internal/middlewares"
	"github.com/go-chi/chi/v5"
)

func CreateRouter(cfg *config.Config) chi.Router {
	router := chi.NewRouter()

	router.Post(
		"/", middlewares.Conveyor(
			func(w http.ResponseWriter, r *http.Request) {
				handlers.CreateShortURL(w, r, cfg)
			},
			middlewares.GzipMiddleware,
			middlewares.WithLogging,
		))
	router.Get("/{id}", middlewares.Conveyor(
		handlers.GetFullURL,
		middlewares.GzipMiddleware,
		middlewares.WithLogging))
	router.Route("/api", func(r chi.Router) {
		r.Post("/shorten", middlewares.Conveyor(
			func(w http.ResponseWriter, r *http.Request) {
				handlers.APICreateShortURL(w, r, cfg)
			},
			middlewares.GzipMiddleware,
			middlewares.WithLogging,
		))
	})

	return router
}
