package app

import (
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func CreateRouter(cfg *config.Config) chi.Router {
	router := chi.NewRouter()

	router.Post("/",
		func(w http.ResponseWriter, r *http.Request) {
			handlers.CreateShortURL(w, r, cfg)
		})
	router.Get("/{id}", handlers.GetFullURL)

	return router
}
