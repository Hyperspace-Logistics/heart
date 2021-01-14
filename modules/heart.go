package modules

import (
	"sync"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
)

var (
	once sync.Once
)

// LoadHeart preloads the heart module for use in the server
func LoadHeart(app *fiber.App, state *lua.State) error {
	state.Register("_static", func(state *lua.State) int {
		route := state.ToString(state.GetTop() - 1)
		filepath := state.ToString(state.GetTop())

		once.Do(func() {
			app.Static(route, filepath)
		})

		return 0
	})

	return state.DoString(`
		package.preload['heart.v1'] = function() 
			local heart = { routes = {}, ctx = require('heart.v1.context') }

			function registerCallback(method, path, callback)
				if heart.routes[path] == nil then
					heart.routes[path] = {}
				end

				heart.routes[path][method] = callback
			end

			function heart.get(path, callback)
				registerCallback("get", path, callback)
			end

			function heart.head(path, callback)
				registerCallback("head", path, callback)
			end

			function heart.post(path, callback)
				registerCallback("post", path, callback)
			end

			function heart.put(path, callback)
				registerCallback("put", path, callback)
			end

			function heart.delete(path, callback)
				registerCallback("delete", path, callback)
			end

			function heart.options(path, callback)
				registerCallback("options", path, callback)
			end

			function heart.trace(path, callback)
				registerCallback("trace", path, callback)
			end

			function heart.patch(path, callback)
				registerCallback("patch", path, callback)
			end

			function heart.static(route, filepath)
				_static(route, filepath)
			end

			function heart.notfound(callback)
				registerCallback('_not_found', '/', callback)
			end

			return heart
		end
	`)
}
