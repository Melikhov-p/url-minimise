package storage

import "errors"

type StorageType int

const (
	BaseStorage StorageType = iota
	StorageFromFile
	StorageInDatabase
)

var errNotFound error = errors.New("not found full url")
