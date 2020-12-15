package main

import (
	"log"
	"os"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/sosodev/heart/core"
	"github.com/sosodev/heart/modules"
	"github.com/sosodev/heart/pool"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: heart [path]")
	}
	path := os.Args[1]

	// TODO: env variable to configure initial pool size
	// maybe make a config struct to deal with env stuff with reasonable defaults
	statePool, err := pool.New(32, func(nuState *lua.State) error {
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

	// TODO: env variable for port
	log.Println("Starting Heart v0.1 💜")
	log.Fatal(app.Listen(":3333"))
}
