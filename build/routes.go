package build

import (
	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/sosodev/heart/pool"
)

// BuildRoutes for the *fiber.App from the initial *lua.State
func Routes(app *fiber.App, state *lua.State, statePool *pool.Pool) {
	state.GetGlobal("app")
	state.GetField(state.GetTop(), "routes")
	state.PushNil()

	for state.Next(-2) != 0 {
		route := state.ToString(-2)

		state.PushNil()
		for state.Next(-2) != 0 {
			method := state.ToString(-2)

			switch method {
			case "get":
				app.Get(route, func(ctx *fiber.Ctx) error {
					reqState, err := statePool.Take()
					if err != nil {
						return err
					}
					defer statePool.Return(reqState)

					reqState.GetGlobal("app")
					reqState.GetField(reqState.GetTop(), "routes")
					reqState.GetField(reqState.GetTop(), route)
					reqState.GetField(reqState.GetTop(), "get")

					reqState.PushNil()
					err = reqState.Call(1, 1)
					if err != nil {
						return err
					}

					response := reqState.ToString(reqState.GetTop())
					reqState.Pop(4)

					return ctx.SendString(response)
				})
			}

			state.Pop(1)
		}

		state.Pop(1)
	}

	state.Pop(2)
}
