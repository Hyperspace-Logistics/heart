package main

import (
	"log"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/sosodev/heart/build"
	"github.com/sosodev/heart/config"
	"github.com/sosodev/heart/modules"
	"github.com/sosodev/heart/pool"
)

func main() {
	config := config.NewConfig()
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	statePool, err := pool.New(config, func(nuState *lua.State) error {
		// TODO: considering reducing lib availibility in Lua
		nuState.OpenLibs()

		err := modules.LoadContext(nuState)
		if err != nil {
			return err
		}

		err = modules.LoadHeart(nuState)
		if err != nil {
			return err
		}

		return nuState.DoFile(config.Path)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer statePool.Cleanup()

	// This function grabs one of the initialStates from the pool to build up the fiber routes
	// It's worth noting that this means that app routes can't be built up dynamically
	// But that's probably not a good idea anyway and implementing it would probably kill performance or me :(
	build.Routes(app, statePool)

	log.Println("Heart v" + config.Version + " is listening to port " + config.Port + " 💜")
	log.Fatal(app.Listen(":" + config.Port))
}
