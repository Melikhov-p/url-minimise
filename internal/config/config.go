package config

import (
	"flag"
	"os"

	storageConfig "github.com/Melikhov-p/url-minimise/internal/repository/config"
	fileConfig "github.com/Melikhov-p/url-minimise/internal/repository/file/config"
	memoryConfig "github.com/Melikhov-p/url-minimise/internal/repository/memory/config"
)

const (
	defaultSrvAddr         = "localhost:8080"
	defaultResAddr         = "http://localhost:8080"
	defaultFileStoragePath = "storage.txt"
	defaultShortURLSize    = 10
)

type Config struct {
	Storage      storageConfig.Config
	ServerAddr   string
	ResultAddr   string
	ShortURLSize int
}

func NewConfig() *Config {
	return &Config{
		ServerAddr: defaultSrvAddr,
		ResultAddr: defaultResAddr,
		Storage: storageConfig.Config{
			InMemory: &memoryConfig.Config{},
			FileStorage: &fileConfig.Config{
				FilePath: defaultFileStoragePath,
			},
		},
		ShortURLSize: defaultShortURLSize,
	}
}

func (c *Config) Build() {
	flag.StringVar(&c.ServerAddr, "a", defaultSrvAddr, "Server host and port")
	flag.StringVar(&c.ResultAddr, "b", defaultResAddr, "Result host and port")
	flag.StringVar(&c.Storage.FileStorage.FilePath, "f", defaultFileStoragePath, "File storage path")
	flag.Parse()

	var (
		srvEnvAddr         string
		resEnvAddr         string
		fileStoragePathEnv string
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
}
