package kv

import (
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/sosodev/heart/config"
)

// KV store for both in-memory and on-disk usage
type KV struct {
	db              *badger.DB
	transaction     *badger.Txn
	transactionLock sync.Mutex
}

var (
	memoryKV *KV
	diskKV   *KV
)

// GetMemoryStore does what it says on the tin
func GetMemoryStore() (*KV, error) {
	if memoryKV == nil {
		memoryDB, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
		if err != nil {
			return nil, err
		}

		memoryKV = &KV{
			db: memoryDB,
		}
	}

	return memoryKV, nil
}

// GetDiskStore does what it says on the tin
func GetDiskStore() (*KV, error) {
	if diskKV == nil {
		diskDB, err := badger.Open(badger.DefaultOptions(config.NewConfig().DBPath).WithSyncWrites(false))
		if err != nil {
			return nil, err
		}

		diskKV = &KV{
			db: diskDB,
		}
	}

	return diskKV, nil
}

// Get the value for the given key without a transcation or error
func (kv *KV) Get(key string) (string, error) {
	var value string
	err := kv.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}

			return fmt.Errorf("failed to read key: %s", err)
		}

		err = item.Value(func(val []byte) error {
			value = string(val)
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to retrieve item value: %s", err)
		}

		return nil
	})

	return value, err
}

// StartTransaction or error
func (kv *KV) StartTransaction() error {
	kv.transactionLock.Lock()
	kv.transaction = kv.db.NewTransaction(true)

	return nil
}

// EndTransaction or error
func (kv *KV) EndTransaction() error {
	err := kv.transaction.Commit()
	if err != nil {
		return err
	}

	kv.transactionLock.Unlock()
	return nil
}

// TransactionGet the value for the given key or error as part of a transaction
func (kv *KV) TransactionGet(key string) (string, error) {
	var value string

	item, err := kv.transaction.Get([]byte(key))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return "", nil
		}

		return "", fmt.Errorf("failed to retrieve item: %s", err)
	}

	err = item.Value(func(val []byte) error {
		value = string(val)
		return nil
	})

	return value, err
}

// TransactionSet the KV pair or error as part of a transaction
func (kv *KV) TransactionSet(key, value string) error {
	err := kv.transaction.Set([]byte(key), []byte(value))
	if err != nil {
		return err
	}

	return nil
}
