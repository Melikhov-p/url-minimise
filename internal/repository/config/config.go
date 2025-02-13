package config

import (
	databaseConfig "github.com/Melikhov-p/url-minimise/internal/repository/database/config"
	fileConfig "github.com/Melikhov-p/url-minimise/internal/repository/file/config"
	memoryConfig "github.com/Melikhov-p/url-minimise/internal/repository/memory/config"
)

// Config конфиг
type Config struct {
	InMemory    *memoryConfig.Config
	FileStorage *fileConfig.Config
	Database    *databaseConfig.DBConfig
}
