package main

import (
	"fmt"
	"github.com/Melikhov-p/url-minimise/internal/app"
	"github.com/Melikhov-p/url-minimise/internal/config"
	"net/http"
)

func main() {
	config.ParseFlags()

	router := app.CreateRouter()

	err := http.ListenAndServe(config.ServerAddr, router)

	if err != nil {
		panic(fmt.Sprintf(`Internal Error %v`, err.Error()))
	}
}
