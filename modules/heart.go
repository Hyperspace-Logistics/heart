package modules

import (
	"sync"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"

	_ "embed"
)

var (
	//go:embed heart.lua
	heartLua string
	once     sync.Once
)

// LoadHeart preloads the heart module for use in the server
func LoadHeart(app *fiber.App, state *lua.State) error {
	state.Register("_static", func(state *lua.State) int {
		once.Do(func() {
			route := state.ToString(state.GetTop() - 1)
			filepath := state.ToString(state.GetTop())
			app.Static(route, filepath)
		})

		return 0
	})

	return state.DoString(heartLua)
}
