package build

import (
	"log"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/sosodev/heart/pool"
)

// Routes for the *fiber.App from the initial *lua.State
//
// TODO:
// * Move switch sub-statements to a function that accepts the method (could be anonymous :o)
// * Break out loop functionality for clarity?
// * Take a closer look at error handling
// *
//
func Routes(app *fiber.App, statePool *pool.Pool) {
	state, err := statePool.Take()
	if err != nil {
		log.Fatal(err)
	}
	defer statePool.Return(state)

	loopRoutes(state, func(route string) {
		state.PushNil()
		defer state.Pop(1)

		for state.Next(-2) != 0 {
			method := state.ToString(-2)

			switch method {
			case "get":
				app.Get(route, func(ctx *fiber.Ctx) error {
					return handleRequest(ctx, "get", route, statePool)
				})
			}

			state.Pop(1)
		}
	})
}

// handle an incoming request with Lua
func handleRequest(ctx *fiber.Ctx, method string, route string, statePool *pool.Pool) error {
	reqState, err := statePool.Take()
	if err != nil {
		return err
	}
	defer statePool.Return(reqState)

	// Bind the *fiber.Ctx to a global
	reqState.PushGoStruct(ctx)
	reqState.SetGlobal("_fiber_ctx")

	// load the callback
	reqState.GetGlobal("app")
	reqState.GetField(reqState.GetTop(), "routes")
	reqState.GetField(reqState.GetTop(), route)
	reqState.GetField(reqState.GetTop(), method)

	// load the context module as the only argument
	// look, I know what you're thinking
	// "Wait a minute... That context module could be required in the user code, right?"
	// To which I would say: well yes, of course it could be and that would probably be more efficient
	// But I want the API to look pretty and that would be kind of off putting for people who come from a more traditional web server
	// Heart is kind of unique in the way that it could seemingly bind global state to a parallel request
	// and that's just a little weird when our brains are wired to think statelessly I think 🤷
	reqState.GetGlobal("ctx")
	err = reqState.Call(1, 1)
	if err != nil {
		return err
	}

	response := reqState.ToString(reqState.GetTop())
	reqState.Pop(4) // normally I'd defer this pop closer to the stack growth but I've found it makes debugging hard
	return ctx.SendString(response)
}

// loop the routes built up in the app global variable
func loopRoutes(state *lua.State, callback func(string)) {
	state.GetGlobal("app")
	state.GetField(state.GetTop(), "routes")
	state.PushNil()

	for state.Next(-2) != 0 {
		route := state.ToString(-2)
		callback(route)
	}

	state.Pop(2)
}
