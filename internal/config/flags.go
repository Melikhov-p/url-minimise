package config

import (
	"flag"
)

const (
	defaultSrvAddr = "localhost:8080"
	defaultResAddr
)

var (
	ServerAddr string
	ResultAddr string
)

func ParseFlags() {
	flag.StringVar(&ServerAddr, "a", defaultSrvAddr, "Server host and port")
	flag.StringVar(&ResultAddr, "b", defaultResAddr, "Result host and port")
	flag.Parse()
}
