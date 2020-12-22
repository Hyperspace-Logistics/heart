package modules

import (
	"fmt"
	"log"
	"sync"

	"github.com/aarzilli/golua/lua"
	"github.com/dgraph-io/badger/v2"
	"github.com/sosodev/heart/config"
)

var (
	badgerDB            *badger.DB
	transaction         *badger.Txn
	diskTransactionLock sync.Mutex
)

func init() {
	var err error
	badgerDB, err = badger.Open(badger.DefaultOptions(config.NewConfig().DBPath))
	if err != nil {
		log.Fatal(err)
	}
}

// LoadKVDisk modules into Lua
func LoadKVDisk(state *lua.State) error {
	// note: not actually transactionless but it is a read-only transaction
	// which is roughly equivalent to the memory version
	state.Register("_transactionless_disk_get", func(state *lua.State) int {
		key := state.ToBytes(state.GetTop() - 1)

		var value string
		err := badgerDB.View(func(txn *badger.Txn) error {
			item, err := txn.Get(key)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					value = ""
					return nil
				}

				return fmt.Errorf("failed to read key: %s", err)
			}

			err = item.Value(func(val []byte) error {
				value = string(val)
				return nil
			})
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			log.Printf("failed to read key '%s' from disk: %s\n", key, err)
		}

		state.PushString(value)
		return 1
	})

	state.Register("_disk_get", func(state *lua.State) int {
		key := state.ToBytes(state.GetTop())

		item, err := transaction.Get(key)
		if err != nil {
			log.Printf("failed to get key '%s' from disk: %s\n", key, err)
			state.PushString("")
			return 1
		}

		err = item.Value(func(val []byte) error {
			state.PushString(string(val))
			return nil
		})
		if err != nil {
			log.Printf("failed to retrieve disk value from key '%s': %s", key, err)
		}

		return 1
	})

	state.Register("_disk_set", func(state *lua.State) int {
		key := state.ToBytes(state.GetTop() - 1)
		value := state.ToBytes(state.GetTop())

		err := transaction.Set(key, value)
		if err != nil {
			log.Printf("failed to set key '%s' on disk with value '%s': %s\n", key, value, err)
		}

		return 0
	})

	state.Register("_start_disk_transaction", func(state *lua.State) int {
		diskTransactionLock.Lock()
		transaction = badgerDB.NewTransaction(true)
		return 0
	})

	state.Register("_end_disk_transaction", func(state *lua.State) int {
		err := transaction.Commit()
		if err != nil {
			log.Printf("failed to commit disk transaction: %s\n", err)
		}

		diskTransactionLock.Unlock()
		return 0
	})

	return state.DoString(`
		package.preload['heart.v1.kv.disk'] = function()
			local kv = {}
			local store = {}

			function kv.get(key)
				return _transactionless_disk_get(key)
			end

			function store.get(key)
				return _disk_get(key)
			end

			function store.set(key, value)
				return _disk_set(key, value)
			end

			function kv.transaction(callback)
				_start_disk_transaction()
				callback(store)
				_end_disk_transaction()
			end

			return kv
		end
	`)
}
