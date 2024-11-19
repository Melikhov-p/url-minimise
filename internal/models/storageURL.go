package models

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"go.uber.org/zap"
)

const (
	shortURLSize = 10
)

type StorageType string

var (
	BASESTORAGE     StorageType = "base"
	STORAGEFROMFILE StorageType = "from file"
)

type IStorage interface {
	AddURL(*StorageURL)
	GetFullURL(string) string
	GetDB() map[string]*StorageURL
}

type IStorageSaver interface {
	Save(*StorageURL) error
}

type BaseStorage struct {
	db map[string]*StorageURL
}

func (s *BaseStorage) AddURL(newURL *StorageURL) {
	s.db[newURL.ShortURL] = newURL
}

func (s *BaseStorage) GetFullURL(shortURL string) string {
	searchedElem := s.db[shortURL]
	if searchedElem != nil {
		return searchedElem.OriginalURL
	}
	return ""
}

func (s *BaseStorage) GetDB() map[string]*StorageURL {
	return s.db
}

type FileStorage struct {
	BaseStorage
	file    *os.File
	encoder *json.Encoder
	scanner *bufio.Scanner
}

func (s *FileStorage) Save(record *StorageURL) error {
	if err := s.encoder.Encode(record); err != nil {
		return err
	}
	return nil
}

func NewStorage(storageType StorageType, cfg *config.Config, logger *zap.Logger) (IStorage, error) {
	logger.Debug("creating storage with type", zap.String("Type", string(storageType)))
	switch storageType {
	case BASESTORAGE:
		return &BaseStorage{map[string]*StorageURL{}}, nil
	case STORAGEFROMFILE:
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		scan := bufio.NewScanner(file)

		storage := &FileStorage{
			BaseStorage: BaseStorage{
				db: map[string]*StorageURL{},
			},
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

		return storage, nil
	}

	err := errors.New(string(storageType))
	return nil, fmt.Errorf("unknow type of storage %v", err)
}

type StorageURL struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewStorageURL(fullURL string, s IStorage) (*StorageURL, error) {
	short, err := randomString(shortURLSize, s)

	if err == nil {
		return &StorageURL{
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
