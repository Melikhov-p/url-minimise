package storage

import "errors"

type StorageType int

const (
	BaseStorage StorageType = iota
	StorageFromFile
	StorageInDatabase
)

type MarkDeleteResult struct {
	URL string
	Res bool
	Err error
}

var ErrNotFound error = errors.New("not found full url")
var ErrOriginalURLExist error = errors.New("original url already exist in storage")
