package modules

import (
	"log"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/sosodev/heart/las"
)

// LoadContext creates a module for request context
// Functionally it just provides cute bindings to some go functions that can appropriately bridge lua<->fiber
func LoadContext(state *lua.State) error {
	ctx := func(state *lua.State) *fiber.Ctx {
		as, ok := las.Get(state)
		if !ok {
			log.Fatal("failed to load *las.AssociatedState for request")
		}

		return as.Ctx
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

	state.Register("_body", func(state *lua.State) int {
		state.PushString(string(ctx(state).Body()))
		return 1
	})

	state.Register("_set_status", func(state *lua.State) int {
		ctx(state).Status(state.ToInteger(state.GetTop()))
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
				_redirect(path, status)
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
				return context
			end

			function context.status(code)
				_set_status(code)
				return context
			end

			-- converts the given table to a JSON string and returns it
			-- also sets the Content-Type header to application/json
			function context.json(table)
				_set("Content-Type", "application/json")
				return json.encode(table)
			end

			function context.body()
				local body = { value = _body() }

				function body.string()
					return body.value
				end

				function body.json()
					if body.value == '' then
						return {}
					end

					local success, decoded = unsafe_pcall(json.decode, body.value)

					if not success then
						return {}
					end

					return decoded
				end

				return body
			end

			return context
		end
	`)
}
