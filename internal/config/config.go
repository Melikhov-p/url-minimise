package config

import (
	"flag"
	"os"
)

const (
	defaultSrvAddr = "localhost:8080"
	defaultResAddr = "http://localhost:8080"
)

type Config struct {
	ServerAddr string
	ResultAddr string
}

func NewConfig() *Config {
	return &Config{
		ServerAddr: defaultSrvAddr,
		ResultAddr: defaultResAddr,
	}
}

func (c *Config) Build() {
	flag.StringVar(&c.ServerAddr, "a", defaultSrvAddr, "Server host and port")
	flag.StringVar(&c.ResultAddr, "b", defaultResAddr, "Result host and port")
	flag.Parse()

	srvEnvAddr, ok := os.LookupEnv("SERVER_ADDRESS")
	if ok {
		c.ServerAddr = srvEnvAddr
	}

	resEnvAddr, ok := os.LookupEnv("BASE_URL")
	if ok {
		c.ResultAddr = resEnvAddr
	}
}
