package config

import (
	"flag"
	"os"
)

const (
	defaultSrvAddr = "localhost:8080"
	defaultResAddr = "http://localhost:8080"
)

var (
	srvFlagAddr string
	resFlagAddr string
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
	srvEnvAddr, ok := os.LookupEnv("SERVER_ADDRESS")
	resEnvAddr, resEnvExist := os.LookupEnv("BASE_URL")

	c.ServerAddr, c.ResultAddr = srvEnvAddr, resEnvAddr

	flag.StringVar(&srvFlagAddr, "a", defaultSrvAddr, "Server host and port")
	flag.StringVar(&resFlagAddr, "b", defaultResAddr, "Result host and port")
	flag.Parse()

	if !ok {
		c.ServerAddr = srvFlagAddr
	}
	if !resEnvExist {
		c.ResultAddr = resFlagAddr
	}
}
