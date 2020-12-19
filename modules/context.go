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
	ctx := func(state *lua.State) *fiber.Ctx {
		fctx, ok := build.ContextMap.Load(state)
		if !ok {
			log.Fatal("failed to load *fiber.Ctx for request")
		}

		return fctx.(*fiber.Ctx)
	}

	state.Register("_redirect", func(state *lua.State) int {
		path := state.ToString(state.GetTop() - 1)
		code := state.ToInteger(state.GetTop())

		err := ctx(state).Redirect(path, code)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

	state.Register("_pathParams", func(state *lua.State) int {
		key := state.ToString(state.GetTop())
		state.PushString(ctx(state).Params(key))
		return 1
	})

	state.Register("_path", func(state *lua.State) int {
		state.PushString(ctx(state).Path())
		return 1
	})

	state.Register("_set", func(state *lua.State) int {
		key := state.ToString(state.GetTop() - 1)
		value := state.ToString(state.GetTop())

		ctx(state).Set(key, value)

		return 0
	})

	return state.DoString(`
		package.preload['heart.v1.context'] = function()
			local context = {}
			local json = require('heart.v1.json')

			-- redirect to the given path
			-- status is optional and defaults to 302
			function context.redirect(path, status)
				status = status or 302

				_redirect(path, code)
			end

			-- get the value of the given path key
			function context.pathParams(key)
				return _pathParams(key)
			end

			-- get the request path
			function context.path()
				return _path()
			end

			-- set a header with the given K/V
			function context.set(key, value)
				_set(key, value)
			end

			-- converts the given table to a JSON string and returns it
			-- also sets the Content-Type header to application/json
			function context.json(table)
				_set("Content-Type", "application/json")
				return json.encode(table)
			end

			return context
		end
	`)
}