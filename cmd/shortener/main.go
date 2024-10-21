package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const (
	shortURLSize = 10
	host         = "http://localhost:8080/"
)

var AllURLs []FullShortURL

type FullShortURL struct {
	FullURL  string
	ShortURL string
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", GetFullURL)
	mux.HandleFunc("/", CreateShortURL)

	err := http.ListenAndServe(`localhost:8080`, mux)

	if err != nil {
		panic(fmt.Sprintf(`Internal Error %v`, err.Error()))
	}
}

func (fsu *FullShortURL) CheckShortURL(shortURL string) bool {
	return shortURL == fsu.ShortURL
}

func NewRandomString(size int) string { // Создает рандомную строку заданного размера
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	b := make([]rune, size)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}

	return string(b)
}

func CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		fullURL, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `Upal`)
		}

		fullShortURL := FullShortURL{
			FullURL:  string(fullURL),
			ShortURL: NewRandomString(shortURLSize),
		}
		AllURLs = append(AllURLs, fullShortURL)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `%s%s`, host, fullShortURL.ShortURL)
	}
}

func GetFullURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		id := r.PathValue("id")
		var matchURL *FullShortURL

		for _, el := range AllURLs {
			if el.CheckShortURL(id) {
				matchURL = &el
				break
			}
		}

		if matchURL != nil {
			w.Header().Set(`Location`, (*matchURL).FullURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
