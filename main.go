package main

import (
	"log"
	"os"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/sosodev/heart/build"
	"github.com/sosodev/heart/config"
	"github.com/sosodev/heart/modules"
	"github.com/sosodev/heart/pool"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: heart [path]")
	}
	path := os.Args[1]
	config := config.NewConfig()

	statePool, err := pool.New(config.InitialPoolSize, func(nuState *lua.State) error {
		nuState.OpenLibs()

		err := modules.LoadHeart(nuState)
		if err != nil {
			return err
		}

		return nuState.DoFile(path)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer statePool.Cleanup()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Grab one of the initial states to build up the actual app BuildRoutes
	//
	// Note:
	// This does mean that Lua cannot dynamically change framework stuff but I think that's okay
	// The alternative would probably kill performance or kill me (with bugs)
	state, err := statePool.Take()
	if err != nil {
		log.Fatal(err)
	}
	build.Routes(app, state, statePool)
	statePool.Return(state)

	log.Println("Heart v" + config.Version + " is listening to port " + config.Port + " 💜")
	log.Fatal(app.Listen(":" + config.Port))
}
