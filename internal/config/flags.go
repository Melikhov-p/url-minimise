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
	ServerAddr string
	ResultAddr string
)

var (
	srvFlagAddr string
	resFlagAddr string
)

func ParseFlags() {
	srvEnvAddr, srvEnvExist := os.LookupEnv("SERVER_ADDRESS")
	resEnvAddr, resEnvExist := os.LookupEnv("BASE_URL")

	ServerAddr, ResultAddr = srvEnvAddr, resEnvAddr

	flag.StringVar(&srvFlagAddr, "a", defaultSrvAddr, "Server host and port")
	flag.StringVar(&resFlagAddr, "b", defaultResAddr, "Result host and port")
	flag.Parse()

	if !srvEnvExist {
		ServerAddr = srvFlagAddr
	}
	if !resEnvExist {
		ResultAddr = resFlagAddr
	}
}
