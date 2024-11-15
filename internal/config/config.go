package config

import (
	"flag"
	"os"
)

const (
	defaultSrvAddr         = "localhost:8080"
	defaultResAddr         = "http://localhost:8080"
	defaultFileStoragePath = "storage.txt"
)

type Config struct {
	ServerAddr      string
	ResultAddr      string
	FileStoragePath string
}

func NewConfig() *Config {
	return &Config{
		ServerAddr:      defaultSrvAddr,
		ResultAddr:      defaultResAddr,
		FileStoragePath: defaultFileStoragePath,
	}
}

func (c *Config) Build() {
	flag.StringVar(&c.ServerAddr, "a", defaultSrvAddr, "Server host and port")
	flag.StringVar(&c.ResultAddr, "b", defaultResAddr, "Result host and port")
	flag.StringVar(&c.FileStoragePath, "f", defaultFileStoragePath, "File storage path")
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
		c.FileStoragePath = fileStoragePathEnv
	}

}
