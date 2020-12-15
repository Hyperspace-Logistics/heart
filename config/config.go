package config

import (
	"log"
	"os"
	"strconv"
)

// Config for the application
type Config struct {
	Version         string
	InitialPoolSize int
	Port            string
}

// NewConfig gets you a new *Config
func NewConfig() *Config {
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
		Version:         "0.1",
		InitialPoolSize: initialPoolSize,
		Port:            port,
	}
}
