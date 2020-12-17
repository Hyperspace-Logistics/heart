package modules

import (
	"log"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/sosodev/heart/build"
)

// LoadContext creates a module for request context
// Functionally it just provides cute bindings to some go functions that can appropriately bridge lua<->fiber
func LoadContext(state *lua.State) error {
	state.Register("_redirect", redirect)
	state.Register("_pathParams", pathParams)

	return state.DoString(`
		package.preload['heart.v1.context'] = function()
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

func ctx(state *lua.State) *fiber.Ctx {
	fctx, ok := build.ContextMap.Load(state)
	if !ok {
		log.Fatal("missing *fiber.Ctx for request")
	}

	return fctx.(*fiber.Ctx)
}

func redirect(state *lua.State) int {
	path := state.ToString(state.GetTop() - 1)
	code := state.ToInteger(state.GetTop())

	err := ctx(state).Redirect(path, code)
	if err != nil {
		log.Fatal(err)
	}

	return 0
}

// binds to ctx.pathParams to dynamically bring *fiber.Ctx path params into the lua context
func pathParams(state *lua.State) int {
	key := state.ToString(state.GetTop())
	state.PushString(ctx(state).Params(key))
	return 1
}
