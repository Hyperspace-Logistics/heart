package main

import (
	"os"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
	"github.com/sosodev/heart/build"
	"github.com/sosodev/heart/config"
	"github.com/sosodev/heart/modules"
	"github.com/sosodev/heart/pool"
)

func main() {
	// initial logging setup
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("PROD") != "prod" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	config := config.NewConfig()
	zerolog.SetGlobalLevel(config.LogLevel)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// logging middleware
	app.Use(func(c *fiber.Ctx) error {
		// start timing request
		start := time.Now()

		// handle the request
		chainErr := c.Next()
		if chainErr != nil {
			err := app.Config().ErrorHandler(c, chainErr)
			if err != nil {
				c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		responseTime := time.Since(start)
		log.Info().Int("status", c.Response().StatusCode()).Str("method", c.Method()).Str("path", c.Path()).Str("response_time", responseTime.String()).Msg("Request")
		return nil
	})

	statePool, err := pool.New(config, func(nuState *lua.State) error {
		// TODO: considering reducing lib availibility in Lua
		nuState.OpenLibs()

		// Load modules to be used in the Lua code
		// Unfortunately order does matter here
		// Heart depends on context which depends on JSON
		err := modules.LoadJSON(nuState)
		if err != nil {
			return err
		}

		err = modules.LoadContext(nuState)
		if err != nil {
			return err
		}

		err = modules.LoadKV(nuState)
		if err != nil {
			return err
		}

		err = modules.LoadHeart(app, nuState)
		if err != nil {
			return err
		}

		return nuState.DoFile(config.Path)
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize lua state")
	}
	defer statePool.Cleanup()

	// This function grabs one of the initialStates from the pool to build up the fiber routes
	// It's worth noting that this means that app routes can't be built up dynamically
	// But that's probably not a good idea anyway and implementing it would probably kill performance or me :(
	build.Routes(app, statePool)

	// swap out the log's writer for a non-blocking one
	// this greatly increases logging throughput
	nonBlockingWriter := diode.NewWriter(os.Stdout, 1000, 0, func(missed int) {})
	defer nonBlockingWriter.Close()
	if config.Production {
		log.Logger = log.Output(nonBlockingWriter)
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: nonBlockingWriter})
	}

	log.Info().Str("port", config.Port).Msg("Heart is online 💜")
	log.Fatal().Err(app.Listen(":" + config.Port)).Msg("App failed to run")
}
