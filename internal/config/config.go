package config

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"time"

	storageConfig "github.com/Melikhov-p/url-minimise/internal/repository/config"
	databaseConfig "github.com/Melikhov-p/url-minimise/internal/repository/database/config"
	fileConfig "github.com/Melikhov-p/url-minimise/internal/repository/file/config"
	memoryConfig "github.com/Melikhov-p/url-minimise/internal/repository/memory/config"
	"github.com/Melikhov-p/url-minimise/internal/storage"
	_ "github.com/jackc/pgx/v5" // PostgreSQL driver.
	"go.uber.org/zap"
)

const (
	defaultSrvAddr         = "localhost:8080"
	defaultResAddr         = "http://localhost:8080"
	defaultFileStoragePath = "storage.txt"
	defaultMigrationsPath  = "./internal/storage/migrations"
	defaultShortURLSize    = 10
	defaultTLS             = false
	defaultStorageMode     = storage.StorageFromFile
)

// cfgFromFile structure for fields from config file.
type cfgFromFile struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDsn     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// Config структура конфига.
type Config struct {
	StorageMode      storage.StorageType
	Storage          storageConfig.Config
	JWTTokenLifeTime time.Duration
	TLS              bool
	ShortURLSize     int
	ServerAddr       string
	ResultAddr       string
	SecretKey        string
	ConfigPath       string
}

// NewConfig Возвращает указатель на конфиг, withoutFlags нужен для тестов, чтобы не читать флаги постоянно.
func NewConfig(logger *zap.Logger, withoutFlags bool) *Config {
	cfg := &Config{
		ServerAddr:       defaultSrvAddr,
		ResultAddr:       defaultResAddr,
		StorageMode:      defaultStorageMode,
		JWTTokenLifeTime: 24 * time.Hour,
		TLS:              false,
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
		ConfigPath:   "",
	}
	if withoutFlags {
		return cfg
	}

	cfg.build(logger)

	return cfg
}

// getConfigFromFile get config params from file with provided path.
func (c *Config) getConfigFromFile(filePath string, log *zap.Logger) bool {
	if c.ConfigPath == "" {
		log.Debug("path to config file is empty")
		return false
	}

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Error("cannot open config file", zap.String("path", c.ConfigPath), zap.Error(err))
		return false
	}
	defer func() {
		_ = f.Close()
	}()

	log.Debug("config file opened")
	cfgF := cfgFromFile{
		ServerAddress:   "",
		BaseURL:         "",
		FileStoragePath: "",
		DatabaseDsn:     "",
		EnableHTTPS:     false,
	}

	data, err := io.ReadAll(f)
	if err != nil {
		log.Error("error read bytes", zap.Error(err))
		return false
	}

	err = json.Unmarshal(data, &cfgF)
	if err != nil {
		log.Error("error unmarshal", zap.Error(err))
		return false
	}

	log.Debug("config from file has been read")
	if cfgF.EnableHTTPS {
		c.TLS = true
	}
	if cfgF.FileStoragePath != "" && c.Storage.FileStorage.FilePath == defaultFileStoragePath {
		c.Storage.FileStorage.FilePath = cfgF.FileStoragePath
	}
	if cfgF.DatabaseDsn != "" && c.Storage.Database.DSN == "" {
		c.Storage.Database.DSN = cfgF.DatabaseDsn
	}
	if cfgF.BaseURL != "" && c.ResultAddr == defaultResAddr {
		c.ResultAddr = cfgF.BaseURL
	}
	log.Debug("srv addr", zap.String("file", cfgF.ServerAddress), zap.String("struct", c.ServerAddr))
	if cfgF.ServerAddress != "" && c.ServerAddr == defaultSrvAddr {
		c.ServerAddr = cfgF.ServerAddress
	}

	return true
}

func (c *Config) build(logger *zap.Logger) {
	flag.StringVar(&c.ServerAddr, "a", defaultSrvAddr, "Server host and port")
	flag.StringVar(&c.ResultAddr, "b", defaultResAddr, "Result host and port")
	flag.StringVar(&c.Storage.FileStorage.FilePath, "f", defaultFileStoragePath, "File storage path")
	flag.StringVar(&c.Storage.Database.DSN, "d", "", "StorageInDatabase DSN")
	flag.BoolVar(&c.TLS, "s", defaultTLS, "TLS server mode")

	flag.StringVar(&c.ConfigPath, "c", "", "Config file path")
	flag.StringVar(&c.ConfigPath, "config", "", "Config file path")
	flag.Parse()

	c.getConfigFromFile(c.ConfigPath, logger)

	var (
		srvEnvAddr         string
		resEnvAddr         string
		cfgPathEnv         string
		fileStoragePathEnv string
		databaseEnvDSN     string
		ok                 bool
	)

	if cfgPathEnv, ok = os.LookupEnv("CONFIG"); ok {
		c.ConfigPath = cfgPathEnv
	}

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
	if _, ok = os.LookupEnv("ENABLE_HTTPS"); ok {
		c.TLS = true
	}

	if c.Storage.Database.DSN != "" {
		c.StorageMode = storage.StorageInDatabase
		logger.Debug("Database mode ON")
	}
}
