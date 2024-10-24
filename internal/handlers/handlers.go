package handlers

import (
	"fmt"
	"github.com/Melikhov-p/url-minimise/internal/utils"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

const (
	shortURLSize = 10
	host         = "http://localhost:8080/"
)

var AllURLs []FullShortURL //TODO: когда поднимем базу - убрать в отдельный пакет

type FullShortURL struct {
	FullURL  string
	ShortURL string
}

func (fsu *FullShortURL) CheckShortURL(shortURL string) bool {
	return shortURL == fsu.ShortURL
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
			ShortURL: utils.RandomString(shortURLSize),
		}
		AllURLs = append(AllURLs, fullShortURL)

		w.Header().Set(`Content-Type`, `text/plain`)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `%s%s`, host, fullShortURL.ShortURL)
	}
}

func GetFullURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		id := chi.URLParam(r, "id")
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
