package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/go-chi/chi/v5"
)

const (
	shortURLSize = 10
)

type storageURL map[string]string

var shortFullURL storageURL = storageURL{}

func CreateShortURL(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
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
		return
	}

	shortURL, err := randomString(shortURLSize)
	if err != nil {
		log.Printf("error create random string: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortFullURL[shortURL] = string(fullURL)

	w.Header().Set(`Content-Type`, `text/plain`)
	w.WriteHeader(http.StatusCreated)
	_, err = fmt.Fprintf(w, `%s%s`, cfg.ResultAddr+"/", shortURL)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
func randomString(size int) (string, error) { // Создает рандомную строку заданного размера
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	tries := 5 // количество попыток создать уникальную строку

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	for tries > 0 {
		b := make([]rune, size)
		for i := range b {
			b[i] = chars[rnd.Intn(len(chars))]
		}
		str := string(b)

		if ok := checkDuplicates(str); ok {
			return str, nil
		}
		tries--
	}

	return "", errors.New("reached max tries limit")
}
func checkDuplicates(el string) bool {
	checked := shortFullURL[el]
	return checked == ""
}

func GetFullURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := chi.URLParam(r, "id")

	matchURL := shortFullURL[id]
	if matchURL == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	log.Printf("Matched Full URL %v", matchURL)
	w.Header().Set(`Location`, matchURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
