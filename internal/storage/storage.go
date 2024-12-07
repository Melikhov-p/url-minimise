package storage

import "errors"

type StorageType int

const (
	BaseStorage StorageType = iota
	StorageFromFile
	StorageInDatabase
)

var ErrNotFound error = errors.New("not found full url")
var ErrOriginalURLExist error = errors.New("original url already exist in storage")
