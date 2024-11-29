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
var ErrURLExist error = errors.New("url already exist in storage")
