package modules

import (
	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/sosodev/heart/las"
)

// LoadContext creates a module for request context
// Functionally it just provides cute bindings to some go functions that can appropriately bridge lua<->fiber
func LoadContext(state *lua.State) error {
	ctx := func(state *lua.State) *fiber.Ctx {
		as, ok := las.Get(state)
		if !ok {
			log.Fatal().Msg("Failed to load *las.AssociatedState for request")
		}

		return as.Ctx
	}

	state.Register("_redirect", func(state *lua.State) int {
		path := state.ToString(state.GetTop() - 1)
		code := state.ToInteger(state.GetTop())

		err := ctx(state).Redirect(path, code)
		if err != nil {
			log.Error().Err(err).Msg("Failed to redirect")
		}

		return 0
	})

	state.Register("_path_param", func(state *lua.State) int {
		key := state.ToString(state.GetTop())
		state.PushString(ctx(state).Params(key))
		return 1
	})

	state.Register("_form_param", func(state *lua.State) int {
		state.PushString(ctx(state).FormValue(state.ToString(state.GetTop())))
		return 1
	})

	state.Register("_query_param", func(state *lua.State) int {
		state.PushString(ctx(state).Query(state.ToString(state.GetTop())))
		return 1
	})

	state.Register("_path", func(state *lua.State) int {
		state.PushString(ctx(state).Path())
		return 1
	})

	state.Register("_get_header", func(state *lua.State) int {
		state.PushString(ctx(state).Get(state.ToString(state.GetTop())))
		return 1
	})

	state.Register("_set_header", func(state *lua.State) int {
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

	state.Register("_get_cookie", func(state *lua.State) int {
		state.PushString(ctx(state).Cookies(state.ToString(state.GetTop() - 1)))
		return 1
	})

	state.Register("_set_cookie", func(state *lua.State) int {
		ctx(state).Cookie(&fiber.Cookie{
			Name:  state.ToString(state.GetTop() - 1),
			Value: state.ToString(state.GetTop()),
		})
		return 0
	})

	state.Register("_clear_cookie", func(state *lua.State) int {
		ctx(state).ClearCookie(state.ToString(state.GetTop()))
		return 0
	})

	state.Register("_clear_cookies", func(state *lua.State) int {
		ctx(state).ClearCookie()
		return 0
	})

	state.Register("_host", func(state *lua.State) int {
		state.PushString(ctx(state).Hostname())
		return 1
	})

	state.Register("_ip", func(state *lua.State) int {
		state.PushString(ctx(state).IP())
		return 1
	})

	state.Register("_method", func(state *lua.State) int {
		state.PushString(ctx(state).Method())
		return 1
	})

	state.Register("_path", func(state *lua.State) int {
		state.PushString(ctx(state).Path())
		return 1
	})

	state.Register("_protocol", func(state *lua.State) int {
		state.PushString(ctx(state).Protocol())
		return 1
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

			-- get the value of the given path param by key
			function context.pathParam(key)
				return _path_param(key)
			end

			-- get a form value by key
			function context.formParam(key)
				return _form_param(key)
			end

			-- get the value of a query param by key
			function context.queryParam(key)
				return _query_param(key)
			end

			-- get the request path
			function context.path()
				return _path()
			end

			-- get a header by passing a key with no value
			-- set a header by passing a key with a non-nil value
			function context.headers(key, value)
				if value == nil then
					return _get_header(key)
				else
					_set_header(key, value)
				end
			end

			-- get a cookie by passing a key with no value
			-- set a cookie by passing a key with a non-nil value
			function context.cookies(key, value)
				if value == nil then
					return _get_cookie(key)
				else
					_set_cookie(key, value)
				end
			end

			-- clear cookie by key
			function context.clearCookie(key)
				_clear_cookie(key)
			end

			-- clear all cookies
			function context.clearCookies()
				_clear_cookies()
			end

			-- set the response status to the given code
			function context.status(code)
				_set_status(code)
				return context
			end

			-- converts the given table to a JSON string and returns it
			-- also sets the Content-Type header to application/json
			function context.json(table)
				_set_header("Content-Type", "application/json")
				return json.encode(table)
			end

			-- returns a body object that exposes a string() and json() function to get the body in either format
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

			-- get the hostname of the request
			function context.host()
				return _host()
			end

			-- get the ip of the request
			function context.ip()
				return _ip()
			end

			-- get the HTTP method of the request
			function context.method()
				return _method()
			end

			-- get the request protocol
			function context.protocol()
				return _protocol()
			end

			return context
		end
	`)
}
