// Package las is lua associated state
// functionally it's just a module isolating a map of *lua.State to more *AssociatedState
package las

import (
	"sync"

	"github.com/aarzilli/golua/lua"
	"github.com/gofiber/fiber/v2"
	"github.com/sosodev/heart/kv"
)

// AssociatedState is the a collection of state that gets associated with *lua.State
type AssociatedState struct {
	Ctx         *fiber.Ctx
	TakeCount   int
	MemoryStore *kv.KV
	DiskStore   *kv.KV
}

var (
	asm sync.Map
)

// Get the *AssociatedState for the given *lua.State or a false second return value if not found
func Get(state *lua.State) (*AssociatedState, bool) {
	as, ok := asm.Load(state)
	if !ok {
		return nil, ok
	}

	return as.(*AssociatedState), ok
}

// Free the *AssociatedState for the given *lua.State
func Free(state *lua.State) {
	asm.Delete(state)
}

func getOrInit(state *lua.State) *AssociatedState {
	as, ok := Get(state)
	if !ok {
		as = &AssociatedState{}
		asm.Store(state, as)
	}

	return as
}

// Update the associated state for the given state via callback
func Update(state *lua.State, updateFunc func(associatedState *AssociatedState) error) error {
	return updateFunc(getOrInit(state))
}
