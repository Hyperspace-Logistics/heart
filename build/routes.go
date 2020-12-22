package build

import (
	"fmt"
	"log"
	"sync"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/sosodev/heart/pool"
)

var (
	// ContextMap is a mapping of *lua.State to *fiber.Ctx :)
	ContextMap sync.Map
)

// Routes for the *fiber.App from the initial *lua.State
//
// TODO:
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
			handler := func(ctx *fiber.Ctx) error {
				return handleRequest(ctx, method, route, statePool)
			}

			switch method {
			case "get":
				app.Get(route, handler)
			case "head":
				app.Head(route, handler)
			case "post":
				app.Post(route, handler)
			case "put":
				app.Put(route, handler)
			case "delete":
				app.Delete(route, handler)
			case "options":
				app.Options(route, handler)
			case "trace":
				app.Trace(route, handler)
			case "patch":
				app.Patch(route, handler)
			}

			state.Pop(1)
		}
	})
}

// handle an incoming request with Lua
func handleRequest(ctx *fiber.Ctx, method string, route string, statePool *pool.Pool) error {
	reqState, err := statePool.Take()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("500 - Internal Server Error")
	}
	releaseState := false
	defer func() {
		if releaseState {
			reqState.Close()
		} else {
			statePool.Return(reqState)
		}
	}()

	// Update the context map to ensure the *lua.State maps to its current request
	ContextMap.Store(reqState, ctx)

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
	// and that's just a little weird when our brains are wired to think statelessly 🤷
	reqState.GetGlobal("_heart_ctx")
	err = reqState.Call(1, 1)
	if err != nil {
		releaseState = true
		return fmt.Errorf("lua error: %s", err)
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
