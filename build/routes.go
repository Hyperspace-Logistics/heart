package build

import (
	"fmt"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/sosodev/heart/config"
	"github.com/sosodev/heart/las"
	"github.com/sosodev/heart/pool"
)

var (
	appConfig *config.Config = config.NewConfig()
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
		log.Fatal().Err(err).Msg("Failed to retrieve initial lua state")
	}
	defer statePool.Return(state)

	// the 404 handler needs to be registered last
	// so we just take a reference to it when found and register if after everything else
	// this does mean that only a single handler could be used but that's ideal anyway
	var notFoundHandler func(*fiber.Ctx) error

	loopRoutes(state, func(route string) {
		state.PushNil()
		defer state.Pop(1)

		for state.Next(-2) != 0 {
			method := state.ToString(-2)
			handler := func(ctx *fiber.Ctx) error {
				return handleRequest(ctx, method, route, statePool)
			}

			log.Debug().Str("method", method).Str("route", route).Msg("Registering handler")

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
			case "_not_found":
				notFoundHandler = handler
			}

			state.Pop(1)
		}
	})

	// register the 404 handler if found
	if notFoundHandler != nil {
		app.Use(notFoundHandler)
	}
}

// handle an incoming request with Lua
func handleRequest(ctx *fiber.Ctx, method string, route string, statePool *pool.Pool) error {
	reqState, err := statePool.Take()
	if err != nil {
		log.Error().Err(err).Msg("Failed to take request state")

		if appConfig.Production {
			return fmt.Errorf("500 - Internal Server Error")
		}

		return err
	}
	releaseState := false
	defer func() {
		if releaseState {
			reqState.Close()
		} else {
			statePool.Return(reqState)
		}
	}()

	// Associate the *fiber.Ctx with the request *lua.State
	err = las.Update(reqState, func(as *las.AssociatedState) error {
		as.Ctx = ctx
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to update associated state for request")

		if appConfig.Production {
			return fmt.Errorf("500 - Internal Server Error")
		}

		return err
	}

	// load the callback
	reqState.GetGlobal("app")
	initialTop := reqState.GetTop()
	reqState.GetField(initialTop, "routes")
	reqState.GetField(initialTop+1, route)
	reqState.GetField(initialTop+2, method)

	// load the context module as the only argument
	// look, I know what you're thinking
	// "Wait a minute... That context module could be required in the user code, right?"
	// To which I would say: well yes, of course it could be and that would probably be more efficient
	// But I want the API to look pretty and that would be kind of off putting for people who come from a more traditional web server
	// Heart is kind of unique in the way that it could seemingly bind global state to a parallel request
	// and that's just a little weird when our brains are wired to think statelessly 🤷
	reqState.GetField(initialTop, "ctx")
	err = reqState.Call(1, 1)
	if err != nil {
		releaseState = true

		log.Error().Err(err).Msg("Lua failed to handle request")

		if appConfig.Production {
			return fmt.Errorf("500 - Internal Server Error")
		}

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
