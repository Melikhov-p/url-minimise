package main

import (
	"fmt"
	"github.com/Melikhov-p/url-minimise/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	router := chi.NewRouter()

	router.Post("/", handlers.CreateShortURL)
	router.Get("/{id}", handlers.GetFullURL)

	err := http.ListenAndServe(`localhost:8080`, router)

	if err != nil {
		panic(fmt.Sprintf(`Internal Error %v`, err.Error()))
	}
}
