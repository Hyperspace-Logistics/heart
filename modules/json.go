package modules

import (
	"fmt"

	"github.com/aarzilli/golua/lua"

	_ "embed"
)

var (
	//go:embed json.lua
	jsonLua string
)

// LoadJSON module
func LoadJSON(state *lua.State) error {
	return state.DoString(fmt.Sprintf("package.preload['heart.v1.json'] = function() %s end", string(jsonLua)))
}
