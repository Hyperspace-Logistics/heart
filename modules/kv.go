package modules

import (
	"bytes"
	"log"
	"text/template"

	"github.com/aarzilli/golua/lua"
	"github.com/sosodev/heart/kv"
)

// LoadKV modules into Lua
func LoadKV(state *lua.State) error {
	kvGet := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			key := state.ToString(state.GetTop())

			value, err := store.Get(key)
			if err != nil {
				log.Printf("kv.get failed for key '%s': %s", key, err)
				state.PushString("")
				return 1
			}

			state.PushString(value)

			return 1
		}
	}

	storeGet := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			key := state.ToString(state.GetTop())

			value, err := store.TransactionGet(key)
			if err != nil {
				log.Printf("store.get failed for key '%s': %s", key, err)
				state.PushString("")
				return 1
			}

			state.PushString(value)

			return 1
		}
	}

	storeSet := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			key := state.ToString(state.GetTop() - 1)
			value := state.ToString(state.GetTop())

			err := store.TransactionSet(key, value)
			if err != nil {
				log.Printf("store set failed for K/V pair (%s : %s): %s", key, value, err)
			}

			return 0
		}
	}

	startTransaction := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			err := store.StartTransaction()
			if err != nil {
				log.Printf("kv.transaction failed to start: %s", err)
			}

			return 0
		}
	}

	endTransaction := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			err := store.EndTransaction()
			if err != nil {
				log.Printf("kv.transaction failed to commit: %s", err)
			}

			return 0
		}
	}

	memoryStore, err := kv.GetMemoryStore()
	if err != nil {
		return err
	}

	state.Register("_memory_get", kvGet(memoryStore))
	state.Register("_memory_transaction_get", storeGet(memoryStore))
	state.Register("_memory_transaction_set", storeSet(memoryStore))
	state.Register("_start_memory_transaction", startTransaction(memoryStore))
	state.Register("_end_memory_transaction", endTransaction(memoryStore))

	diskStore, err := kv.GetDiskStore()
	if err != nil {
		return err
	}

	state.Register("_disk_get", kvGet(diskStore))
	state.Register("_disk_transaction_get", storeGet(diskStore))
	state.Register("_disk_transaction_set", storeSet(diskStore))
	state.Register("_start_disk_transaction", startTransaction(diskStore))
	state.Register("_end_disk_transaction", endTransaction(diskStore))

	kvTemplate := template.Must(template.New("").Parse(`
		package.preload['heart.v1.kv.{{.medium}}'] = function()
			local kv = {}
			local store = {}

			function kv.get(key)
				return _{{.medium}}_get(key)
			end

			function store.get(key)
				return _{{.medium}}_transaction_get(key)
			end

			function store.set(key, value)
				return _{{.medium}}_transaction_set(key, value)
			end

			function kv.transaction(callback)
				_start_{{.medium}}_transaction()
				callback(store)
				_end_{{.medium}}_transaction()
			end

			return kv
		end
	`))

	diskModule := new(bytes.Buffer)
	err = kvTemplate.Execute(diskModule, map[string]string{"medium": "disk"})
	if err != nil {
		return err
	}

	err = state.DoString(diskModule.String())
	if err != nil {
		return err
	}

	memoryModule := new(bytes.Buffer)
	err = kvTemplate.Execute(memoryModule, map[string]string{"medium": "memory"})
	if err != nil {
		return err
	}

	err = state.DoString(memoryModule.String())
	if err != nil {
		return err
	}

	return nil
}
