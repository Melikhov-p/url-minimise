package storage

import (
	"errors"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

// StorageType тип хранилища.
type StorageType int

// константы
const (
	// BaseStorage базовое хранилище - в памяти.
	BaseStorage StorageType = iota
	// StorageFromFile хранилище в файле.
	StorageFromFile
	// StorageInDatabase хранилище в базе данных.
	StorageInDatabase
)

// MarkDeleteURL адрес отмеченный на удаление.
type MarkDeleteURL struct {
	ShortURL string
	User     *models.User
}

// ErrNotFound полный адрес не найден.
var ErrNotFound error = errors.New("not wantFound full url")

// ErrOriginalURLExist полный адрес уже существует в хранилище.
var ErrOriginalURLExist error = errors.New("original url already exist in storage")
