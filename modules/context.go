package modules

import (
	"log"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
)

// LoadContext creates a module for request context
// Functionally it just provides cute bindings to some go functions that can appropriately bridge lua<->fiber
func LoadContext(state *lua.State) error {
	state.Register("_redirect", redirect)
	state.Register("_pathParams", pathParams)

	return state.DoString(`
		package.preload['heart.context'] = function()
			local context = {}

			function context.redirect(path, code)
				code = code or 302

				_redirect(path, code)
			end

			function context.pathParams(key)
				return _pathParams(key)
			end

			return context
		end
	`)
}

func redirect(state *lua.State) int {
	path := state.ToString(state.GetTop() - 1)
	code := state.ToInteger(state.GetTop())

	state.GetGlobal("_fiber_ctx")
	fiberCtx := state.ToGoStruct(state.GetTop()).(*fiber.Ctx)

	err := fiberCtx.Redirect(path, code)
	if err != nil {
		log.Fatal(err)
	}

	return 0
}

// binds to ctx.pathParams to dynamically bring *fiber.Ctx path params into the lua context
func pathParams(state *lua.State) int {
	key := state.ToString(state.GetTop())
	state.GetGlobal("_fiber_ctx")
	fiberCtx := state.ToGoStruct(state.GetTop()).(*fiber.Ctx)

	state.PushString(fiberCtx.Params(key))
	return 1
}
