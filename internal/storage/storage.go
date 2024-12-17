package storage

import (
	"errors"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

type StorageType int

const (
	BaseStorage StorageType = iota
	StorageFromFile
	StorageInDatabase
)

type MarkDeleteURL struct {
	ShortURL string
	User     *models.User
}

type MarkDeleteResult struct {
	URL string
	Res bool
	Err error
}

var ErrNotFound error = errors.New("not found full url")
var ErrOriginalURLExist error = errors.New("original url already exist in storage")
