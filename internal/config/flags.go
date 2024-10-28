package config

import (
	"flag"
	"log"
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
	log.Printf("Env variables: %s, %s", srvEnvAddr, resEnvAddr)

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
