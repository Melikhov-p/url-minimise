package storage

import "errors"

type StorageType int

const (
	BaseStorage StorageType = iota
	StorageFromFile
	StorageInDatabase
)

const UniqueViolationCode = "23505"

var ErrNotFound error = errors.New("not found full url")

var ErrOriginalURLExist error = errors.New("original url already exist in storage")

var errOriginalURLExist dbError = dbError{
	Field: "url_original_url",
	Code:  UniqueViolationCode,
}

type dbError struct {
	Field string
	Code  string
}
