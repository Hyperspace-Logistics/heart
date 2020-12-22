package modules

import (
	"sync"

	"github.com/aarzilli/golua/lua"
)

var (
	memoryKV              map[string]string = make(map[string]string)
	memoryTransactionLock sync.RWMutex
)

// LoadKVMemory modules into Lua
func LoadKVMemory(state *lua.State) error {
	state.Register("_transactionless_memory_get", func(state *lua.State) int {
		memoryTransactionLock.RLock()
		defer memoryTransactionLock.RUnlock()

		key := state.ToString(state.GetTop())

		state.PushString(memoryKV[key])

		return 1
	})

	state.Register("_memory_get", func(state *lua.State) int {
		key := state.ToString(state.GetTop())

		state.PushString(memoryKV[key])

		return 1
	})

	state.Register("_memory_set", func(state *lua.State) int {
		key := state.ToString(state.GetTop() - 1)
		value := state.ToString(state.GetTop())

		memoryKV[key] = value

		return 0
	})

	state.Register("_start_memory_transaction", func(state *lua.State) int {
		memoryTransactionLock.Lock()
		return 0
	})

	state.Register("_end_memory_transaction", func(state *lua.State) int {
		memoryTransactionLock.Unlock()
		return 0
	})

	return state.DoString(`
		package.preload['heart.v1.kv.memory'] = function()
			local kv = {}
			local store = {}

			function kv.get(key)	
				return _transactionless_memory_get(key)
			end

			function store.get(key)
				return _memory_get(key)
			end

			function store.set(key, value)
				return _memory_set(key, value)
			end

			function kv.transaction(callback)
				_start_memory_transaction()
				callback(store)
				_end_memory_transaction()
			end

			return kv
		end
	`)
}
