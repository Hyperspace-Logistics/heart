package config

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config for the application
type Config struct {
	Production      bool
	Profile         bool
	Path            string
	DBPath          string
	Version         string
	InitialPoolSize int
	Port            string
	DBSyncWrites    bool
	LogLevel        zerolog.Level
}

// NewConfig gets you a new *Config
func NewConfig() *Config {
	if len(os.Args) < 2 {
		log.Fatal().Msg("Usage: heart [path]")
	}
	path := os.Args[1]

	production := os.Getenv("PROD") == "true"
	profile := os.Getenv("PROFILE") == "true"

	dbPath := os.Getenv("DB_PATH")
	if len(dbPath) == 0 {
		dbPath = "./.heart_db"
	}

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3333"
	}

	initialPoolSizeString := os.Getenv("INITIAL_POOL_SIZE")
	if len(initialPoolSizeString) == 0 {
		initialPoolSizeString = "8"
	}

	initialPoolSize, err := strconv.Atoi(initialPoolSizeString)
	if err != nil {
		log.Fatal().Msg("Env variable INITIAL_POOL_SIZE should be an integer")
	}

	dbSyncWrites := os.Getenv("DB_SYNC_WRITES") != "false"

	logLevel := zerolog.InfoLevel
	switch os.Getenv("LOG_LEVEL") {
	case "panic":
		logLevel = zerolog.PanicLevel
	case "fatal":
		logLevel = zerolog.FatalLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "debug":
		logLevel = zerolog.DebugLevel
	case "trace":
		logLevel = zerolog.TraceLevel
	}

	return &Config{
		Production:      production,
		Profile:         profile,
		Path:            path,
		Version:         "0.1",
		InitialPoolSize: initialPoolSize,
		Port:            port,
		DBPath:          dbPath,
		DBSyncWrites:    dbSyncWrites,
		LogLevel:        logLevel,
	}
}
