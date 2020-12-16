package config

import (
	"log"
	"os"
	"strconv"
)

// Config for the application
type Config struct {
	Path            string
	Version         string
	InitialPoolSize int
	Port            string
}

// NewConfig gets you a new *Config
func NewConfig() *Config {
	if len(os.Args) < 2 {
		log.Fatal("usage: heart [path]")
	}
	path := os.Args[1]

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3333"
	}

	initialPoolSizeString := os.Getenv("INITIAL_POOL_SIZE")
	if len(initialPoolSizeString) == 0 {
		initialPoolSizeString = "32"
	}

	initialPoolSize, err := strconv.Atoi(initialPoolSizeString)
	if err != nil {
		log.Fatal("invalid env variable INITIAL_POOL_SIZE: should be an integer")
	}

	return &Config{
		Path:            path,
		Version:         "0.1",
		InitialPoolSize: initialPoolSize,
		Port:            port,
	}
}
