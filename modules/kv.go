package modules

import (
	"bytes"
	"fmt"
	"math/rand"
	"text/template"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"github.com/sosodev/heart/kv"
	"github.com/sosodev/heart/las"
)

// LoadKV modules into Lua
func LoadKV(state *lua.State) error {
	// Associate the databases with the state
	err := las.Update(state, func(as *las.AssociatedState) error {
		memoryStore, err := kv.GetMemoryStore()
		if err != nil {
			return err
		}
		as.MemoryStore = memoryStore

		diskStore, err := kv.GetDiskStore()
		if err != nil {
			return err
		}
		as.DiskStore = diskStore

		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to associate *kv.KV: %s", err)
	}

	associatedStore := func(medium string) *kv.KV {
		switch medium {
		case "disk":
			as, ok := las.Get(state)
			if !ok {
				log.Fatal().Msg("Failed to retrieve disk store from las")
			}
			return as.DiskStore
		case "memory":
			as, ok := las.Get(state)
			if !ok {
				log.Fatal().Msg("Failed to retrieve memory store from las")
			}
			return as.MemoryStore
		}

		log.Fatal().Msg("Incorrect medium for fetching associated store")
		return nil
	}

	kvGet := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			key := state.ToString(state.GetTop())

			value, err := store.Get(key)
			if err != nil {
				log.Error().Str("key", key).Err(err).Msg("kv.get failed")
				state.PushString("")
				return 1
			}

			state.PushString(value)

			return 1
		}
	}

	kvListKeys := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			prefix := state.ToString(state.GetTop() - 1)
			limit := state.ToInteger(state.GetTop())

			results, err := store.ListKeys(prefix, limit)
			if err != nil {
				log.Error().Err(err).Msg("kv.listKeys failed")
			}

			state.NewTable()
			for i, result := range results {
				state.PushString(result)
				state.RawSeti(state.GetTop()-1, i+1)
			}

			return 1
		}
	}

	kvListPairs := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			prefix := state.ToString(state.GetTop() - 1)
			limit := state.ToInteger(state.GetTop())

			results, err := store.ListPairs(prefix, limit)
			if err != nil {
				log.Error().Err(err).Msgf("kv.listPairs failed")
			}

			state.NewTable()
			for i, result := range results {
				state.NewTable()
				state.PushString(result.Key)
				state.SetField(state.GetTop()-1, "key")

				state.PushString(result.Value)
				state.SetField(state.GetTop()-1, "value")

				state.RawSeti(state.GetTop()-1, i+1)
			}

			return 1
		}
	}

	storeGet := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			key := state.ToString(state.GetTop())

			value, err := store.TransactionGet(key)
			if err != nil {
				log.Error().Str("key", key).Err(err).Msgf("store.get failed")
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
				log.Error().Str("key", key).Str("value", value).Err(err).Msg("store.set failed")
			}

			return 0
		}
	}

	storeDelete := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			key := state.ToString(state.GetTop())

			err := store.TransactionDelete(key)
			if err != nil {
				log.Error().Str("key", key).Err(err).Msg("store.delete failed")
			}

			return 0
		}
	}

	startTransaction := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			err := store.StartTransaction()
			if err != nil {
				log.Error().Err(err).Msg("kv.transaction failed to start")
			}

			return 0
		}
	}

	endTransaction := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			err := store.EndTransaction()
			if err != nil {
				log.Error().Err(err).Msg("kv.transaction failed to commit")
			}

			return 0
		}
	}

	startSerialTransaction := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			err := store.StartSerialTransaction()
			if err != nil {
				log.Error().Err(err).Msg("kv.serialTransaction failed to start")
			}

			return 0
		}
	}

	endSerialTransaction := func(store *kv.KV) func(*lua.State) int {
		return func(state *lua.State) int {
			err := store.EndSerialTransaction()
			if err != nil {
				log.Error().Err(err).Msg("kv.serialTransaction failed to commit")
			}

			return 0
		}
	}

	memoryStore := associatedStore("memory")

	state.Register("_memory_get", kvGet(memoryStore))
	state.Register("_memory_list_keys", kvListKeys(memoryStore))
	state.Register("_memory_list_pairs", kvListPairs(memoryStore))
	state.Register("_memory_transaction_get", storeGet(memoryStore))
	state.Register("_memory_transaction_set", storeSet(memoryStore))
	state.Register("_memory_transaction_delete", storeDelete(memoryStore))
	state.Register("_start_memory_transaction", startTransaction(memoryStore))
	state.Register("_end_memory_transaction", endTransaction(memoryStore))
	state.Register("_start_memory_serial_transaction", startSerialTransaction(memoryStore))
	state.Register("_end_memory_serial_transaction", endSerialTransaction(memoryStore))

	diskStore := associatedStore("disk")

	state.Register("_disk_get", kvGet(diskStore))
	state.Register("_disk_list_keys", kvListKeys(diskStore))
	state.Register("_disk_list_pairs", kvListPairs(diskStore))
	state.Register("_disk_transaction_get", storeGet(diskStore))
	state.Register("_disk_transaction_set", storeSet(diskStore))
	state.Register("_disk_transaction_delete", storeDelete(diskStore))
	state.Register("_start_disk_transaction", startTransaction(diskStore))
	state.Register("_end_disk_transaction", endTransaction(diskStore))
	state.Register("_start_disk_serial_transaction", startSerialTransaction(diskStore))
	state.Register("_end_disk_serial_transaction", endSerialTransaction(diskStore))

	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	state.Register("_generate_ulid", func(state *lua.State) int {
		state.PushString(ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String())
		return 1
	})

	kvTemplate := template.Must(template.New("").Parse(`
		package.preload['heart.v1.kv.{{.medium}}'] = function()
			local kv = {}
			local store = {}

			function kv.get(key)
				return _{{.medium}}_get(key)
			end

			function kv.listKeys(prefix, limit)
				return _{{.medium}}_list_keys(prefix, limit)
			end

			function kv.listPairs(prefix, limit)
				return _{{.medium}}_list_pairs(prefix, limit)
			end
			
			function kv.ulid()
				return _generate_ulid()
			end

			function store.get(key)
				return _{{.medium}}_transaction_get(key)
			end

			function store.set(key, value)
				_{{.medium}}_transaction_set(key, value)
			end

			function store.delete(key)
				_{{.medium}}_transaction_delete(key)
			end

			function kv.transaction(callback)
				_start_{{.medium}}_transaction()
				callback(store)
				_end_{{.medium}}_transaction()
			end

			function kv.serialTransaction(callback)
				_start_{{.medium}}_serial_transaction()
				callback(store)
				_end_{{.medium}}_serial_transaction()
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
