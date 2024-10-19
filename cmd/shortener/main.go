package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const (
	shortUrlSize = 10
	host         = "http://localhost:8080/"
)

var AllUrls []FullShortUrl

type FullShortUrl struct {
	FullUrl  string
	ShortUrl string
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", GetFullUrl)
	mux.HandleFunc("/", CreateShortUrl)

	err := http.ListenAndServe(`localhost:8080`, mux)

	if err != nil {
		panic(fmt.Sprintf(`Internal Error %v`, err.Error()))
	}
}

func (fsu *FullShortUrl) CheckShortUrl(shortUrl string) bool {
	if shortUrl == fsu.ShortUrl {
		return true
	}
	return false
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

func CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		fullUrl, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `Upal`)
		}

		fullShortUrl := FullShortUrl{
			FullUrl:  string(fullUrl),
			ShortUrl: NewRandomString(shortUrlSize),
		}
		AllUrls = append(AllUrls, fullShortUrl)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `%s%s`, host, fullShortUrl.ShortUrl)
	}
}

func GetFullUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		id := r.PathValue("id")
		var matchUrl *FullShortUrl

		for _, el := range AllUrls {
			if el.CheckShortUrl(id) {
				matchUrl = &el
				break
			}
		}

		if matchUrl != nil {
			w.Header().Set(`Location`, (*matchUrl).FullUrl)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
