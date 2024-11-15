package models

import (
	"bufio"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"go.uber.org/zap"
)

const (
	shortURLSize = 10
)

type Storage struct {
	db      map[string]*StorageURL
	file    *os.File
	encoder *json.Encoder
	scanner *bufio.Scanner
}

func NewStorageFromFile(cfg *config.Config) (*Storage, error) {
	file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	scan := bufio.NewScanner(file)

	storage := Storage{
		db:      map[string]*StorageURL{},
		file:    file,
		encoder: json.NewEncoder(file),
	}

	var element StorageURL
	for scan.Scan() {
		err = json.Unmarshal(scan.Bytes(), &element)
		if err != nil {
			return nil, err
		}
		storage.db[element.ShortURL] = &element
	}

	return &storage, nil
}

func (s *Storage) Save(record *StorageURL) error {
	if err := s.encoder.Encode(record); err != nil {
		return err
	}
	return nil
}

func (s *Storage) AddURL(newURL *StorageURL) error {
	s.db[newURL.ShortURL] = newURL
	if err := s.Save(newURL); err != nil {
		logger.Log.Error("error saving new URL in storage", zap.Error(err))
		return err
	}
	return nil
}

func (s *Storage) GetFullURL(shortURL string) string {
	searchedElem := s.db[shortURL]
	if searchedElem != nil {
		return searchedElem.OriginalURL
	}
	return ""
}

type StorageURL struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewStorageURL(fullURL string, s *Storage) (*StorageURL, error) {
	short, err := randomString(shortURLSize, s)

	if err == nil {
		return &StorageURL{
			ShortURL:    short,
			OriginalURL: fullURL,
		}, nil
	}
	return nil, err
}

func randomString(size int, s *Storage) (string, error) { // Создает рандомную строку заданного размера
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
func checkDuplicates(el string, s *Storage) bool {
	checked := s.db[el]
	return checked == nil
}
