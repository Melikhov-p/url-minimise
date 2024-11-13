package models

import (
	"errors"
	"math/rand"
	"time"
)

const (
	shortURLSize = 10
)

type storageURL struct {
	URLs newURL
}

type newURL map[string]string

var Storage storageURL = storageURL{URLs: map[string]string{}}

func (s *storageURL) AddURL(fullURL string) (string, error) {
	short, err := randomString(shortURLSize, s)
	if err == nil {
		s.URLs[short] = fullURL
		return short, nil
	}

	return "", err
}

func (s *storageURL) GetFullURL(shortURL string) string {
	return s.URLs[shortURL]
}

func randomString(size int, s *storageURL) (string, error) { // Создает рандомную строку заданного размера
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

		if ok := checkDuplicates(str, s); ok {
			return str, nil
		}
		tries--
	}

	return "", errors.New("reached max tries limit")
}
func checkDuplicates(el string, s *storageURL) bool {
	checked := (*s).URLs[el]
	return checked == ""
}
