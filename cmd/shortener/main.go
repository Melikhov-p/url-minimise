package main

import (
	"fmt"
	"github.com/Melikhov-p/url-minimise/internal/app"
	"github.com/Melikhov-p/url-minimise/internal/config"
	"log"
	"net/http"
)

func main() {
	config.ParseFlags()

	router := app.CreateRouter()

	log.Printf("Running server on %s", config.ServerAddr)
	err := http.ListenAndServe(config.ServerAddr, router)

	if err != nil {
		panic(fmt.Sprintf(`Internal Error %v`, err.Error()))
	}
}
