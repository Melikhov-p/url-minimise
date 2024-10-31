package main

import (
	"github.com/Melikhov-p/url-minimise/internal/app"
	"github.com/Melikhov-p/url-minimise/internal/config"
	"log"
	"net/http"
)

func main() {
	var cfg *config.Config = config.NewConfig()
	cfg.Build()

	router := app.CreateRouter(cfg)

	log.Printf("Running server on %s", cfg.ServerAddr)
	err := http.ListenAndServe(cfg.ServerAddr, router)

	if err != nil {
		log.Fatal(err)
	}
}
