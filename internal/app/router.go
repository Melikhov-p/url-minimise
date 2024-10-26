package app

import (
	"github.com/Melikhov-p/url-minimise/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func CreateRouter() chi.Router {

	router := chi.NewRouter()

	router.Post("/", handlers.CreateShortURL)
	router.Get("/{id}", handlers.GetFullURL)

	return router
}
