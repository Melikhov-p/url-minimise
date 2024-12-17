package config

import (
	"flag"
	"os"
	"time"

	storageConfig "github.com/Melikhov-p/url-minimise/internal/repository/config"
	databaseConfig "github.com/Melikhov-p/url-minimise/internal/repository/database/config"
	fileConfig "github.com/Melikhov-p/url-minimise/internal/repository/file/config"
	memoryConfig "github.com/Melikhov-p/url-minimise/internal/repository/memory/config"
	"github.com/Melikhov-p/url-minimise/internal/storage"
	_ "github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const (
	defaultSrvAddr         = "localhost:8080"
	defaultResAddr         = "http://localhost:8080"
	defaultFileStoragePath = "storage.txt"
	defaultMigrationsPath  = "./internal/storage/migrations"
	defaultShortURLSize    = 10
	defaultStorageMode     = storage.StorageFromFile
)

type Config struct {
	StorageMode      storage.StorageType
	Storage          storageConfig.Config
	JWTTokenLifeTime time.Duration
	ShortURLSize     int
	ServerAddr       string
	ResultAddr       string
	SecretKey        string
}

// NewConfig Возвращает указатель на конфиг, withoutFlags нужен для тестов, чтобы не читать флаги постоянно.
func NewConfig(logger *zap.Logger, withoutFlags bool) *Config {
	cfg := &Config{
		ServerAddr:       defaultSrvAddr,
		ResultAddr:       defaultResAddr,
		StorageMode:      defaultStorageMode,
		JWTTokenLifeTime: 24 * time.Hour,
		Storage: storageConfig.Config{
			InMemory: &memoryConfig.Config{},
			FileStorage: &fileConfig.Config{
				FilePath: defaultFileStoragePath,
			},
			Database: &databaseConfig.DBConfig{
				DSN:            "",
				MigrationsPath: defaultMigrationsPath,
			},
		},
		ShortURLSize: defaultShortURLSize,
		SecretKey:    "",
	}
	if withoutFlags {
		return cfg
	}

	cfg.build(logger)

	return cfg
}

func (c *Config) build(logger *zap.Logger) {
	flag.StringVar(&c.ServerAddr, "a", defaultSrvAddr, "Server host and port")
	flag.StringVar(&c.ResultAddr, "b", defaultResAddr, "Result host and port")
	flag.StringVar(&c.Storage.FileStorage.FilePath, "f", defaultFileStoragePath, "File storage path")
	flag.StringVar(&c.Storage.Database.DSN, "d", "", "StorageInDatabase DSN")
	flag.Parse()

	var (
		srvEnvAddr         string
		resEnvAddr         string
		fileStoragePathEnv string
		databaseEnvDSN     string
		ok                 bool
	)

	if srvEnvAddr, ok = os.LookupEnv("SERVER_ADDRESS"); ok {
		c.ServerAddr = srvEnvAddr
	}
	if resEnvAddr, ok = os.LookupEnv("BASE_URL"); ok {
		c.ResultAddr = resEnvAddr
	}
	if fileStoragePathEnv, ok = os.LookupEnv("FILE_STORAGE_PATH"); ok {
		c.Storage.FileStorage.FilePath = fileStoragePathEnv
	}
	if databaseEnvDSN, ok = os.LookupEnv("DATABASE_DSN"); ok {
		c.Storage.Database.DSN = databaseEnvDSN
	}

	if c.Storage.Database.DSN != "" {
		c.StorageMode = storage.StorageInDatabase
		logger.Debug("Database mode ON")
	}
}
