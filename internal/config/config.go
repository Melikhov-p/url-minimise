package config

import (
	"flag"
	"log"
	"os"

	storageConfig "github.com/Melikhov-p/url-minimise/internal/repository/config"
	databaseConfig "github.com/Melikhov-p/url-minimise/internal/repository/database/config"
	fileConfig "github.com/Melikhov-p/url-minimise/internal/repository/file/config"
	memoryConfig "github.com/Melikhov-p/url-minimise/internal/repository/memory/config"
	"github.com/Melikhov-p/url-minimise/internal/storage"
	_ "github.com/jackc/pgx/v5"
)

const (
	defaultSrvAddr         = "localhost:8080"
	defaultResAddr         = "http://localhost:8080"
	defaultFileStoragePath = "storage.txt"
	defaultShortURLSize    = 10
	defaultStorageMode     = storage.StorageFromFile
)

type Config struct {
	StorageMode  storage.StorageType
	Storage      storageConfig.Config
	ShortURLSize int
	ServerAddr   string
	ResultAddr   string
}

func NewConfig() *Config {
	return &Config{
		ServerAddr:  defaultSrvAddr,
		ResultAddr:  defaultResAddr,
		StorageMode: defaultStorageMode,
		Storage: storageConfig.Config{
			InMemory: &memoryConfig.Config{},
			FileStorage: &fileConfig.Config{
				FilePath: defaultFileStoragePath,
			},
			Database: &databaseConfig.DBConfig{
				DSN: "",
			},
		},
		ShortURLSize: defaultShortURLSize,
	}
}

func (c *Config) Build() {
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
		log.Println("Database mode ON")
	}
}
