package handlers

import (
	"fmt"
	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	shortURLSize = 10
)

var AllURLs []FullShortURL //TODO: когда поднимем базу - убрать в отдельный пакет

type FullShortURL struct {
	FullURL  string
	ShortURL string
}

func (fsu *FullShortURL) checkShortURL(shortURL string) bool {
	return shortURL == fsu.ShortURL
}

func CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fullURL, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
	}

	fullShortURL := FullShortURL{
		FullURL:  string(fullURL),
		ShortURL: randomString(shortURLSize),
	}
	AllURLs = append(AllURLs, fullShortURL)

	w.Header().Set(`Content-Type`, `text/plain`)
	w.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(w, `%s%s`, config.ResultAddr+"/", fullShortURL.ShortURL)
}
func randomString(size int) string { // Создает рандомную строку заданного размера
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

func GetFullURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := chi.URLParam(r, "id")
	var matchURL *FullShortURL

	for _, el := range AllURLs {
		if el.checkShortURL(id) {
			matchURL = &el
			break
		}
	}

	if matchURL != nil {
		w.Header().Set(`Location`, (*matchURL).FullURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
