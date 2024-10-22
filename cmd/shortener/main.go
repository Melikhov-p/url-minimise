package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", GetFullURL)
	mux.HandleFunc("/", CreateShortURL)

	err := http.ListenAndServe(`localhost:8080`, mux)

	if err != nil {
		panic(fmt.Sprintf(`Internal Error %v`, err.Error()))
	}
}
