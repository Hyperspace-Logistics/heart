package modules

import (
	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/sosodev/heart/las"

	_ "embed"
)

var (
	//go:embed context.lua
	contextLua string // embedding the matching lua file for the module
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

	return state.DoString(contextLua)
}
