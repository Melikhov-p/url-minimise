package config

import (
	fileConfig "github.com/Melikhov-p/url-minimise/internal/repository/file/config"
	memoryConfig "github.com/Melikhov-p/url-minimise/internal/repository/memory/config"
)

type Config struct {
	InMemory    *memoryConfig.Config
	FileStorage *fileConfig.Config
}
