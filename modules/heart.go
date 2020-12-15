package modules

import (
	"github.com/aarzilli/golua/lua"
)

// LoadHeart preloads the heart module for use in the server
func LoadHeart(state *lua.State) error {
	return state.DoString(`
		package.preload['heart'] = function() 
			local heart = { routes = {} }

			function heart.get(path, callback)
				if heart.routes[path] == nil then
					heart.routes[path] = {}
				end

				heart.routes[path].get = callback
			end

			return heart
		end
	`)
}
