package repository

import (
	"errors"
	"math/rand"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
)

func NewStorageURL(fullURL string, s IStorage, cfg *config.Config) (*models.StorageURL, error) {
	short, err := randomString(cfg.ShortURLSize, s)

	if err == nil {
		return &models.StorageURL{
			ShortURL:    short,
			OriginalURL: fullURL,
		}, nil
	}
	return nil, err
}

func randomString(size int, s IStorage) (string, error) { // Создает рандомную строку заданного размера
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
func checkDuplicates(el string, s IStorage) bool {
	checked := s.GetDB()[el]
	return checked == nil
}
