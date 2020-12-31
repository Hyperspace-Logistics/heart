package kv

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog/log"
	"github.com/sosodev/heart/config"
)

// KV store for both in-memory and on-disk usage
type KV struct {
	medium      string
	db          *badger.DB
	transaction *badger.Txn
}

// Pair you know
type Pair struct {
	Key   string
	Value string
}

var (
	memoryDB         *badger.DB
	memorySerialLock sync.Mutex
	diskDB           *badger.DB
	diskSerialLock   sync.Mutex
)

// LogWrapper for translating badger logs  to zerolog logs
type LogWrapper struct{}

// Errorf emits a formatted badger error
func (*LogWrapper) Errorf(s string, i ...interface{}) {
	log.Error().Msgf(strings.TrimSuffix(s, "\n"), i...)
}

// Warningf emits a formatted badger warning
func (*LogWrapper) Warningf(s string, i ...interface{}) {
	log.Warn().Msgf(strings.TrimSuffix(s, "\n"), i...)
}

// Infof emits formatted badger info
func (*LogWrapper) Infof(s string, i ...interface{}) {
	log.Info().Msgf(strings.TrimSuffix(s, "\n"), i...)
}

// Debugf emits formatted badger debugging info
func (*LogWrapper) Debugf(s string, i ...interface{}) {
	log.Debug().Msgf(strings.TrimSuffix(s, "\n"), i...)
}

// GetMemoryStore does what it says on the tin
func GetMemoryStore() (*KV, error) {
	var err error
	if memoryDB == nil {
		memoryDB, err = badger.Open(badger.DefaultOptions("").WithInMemory(true).WithLogger(&LogWrapper{}))
		if err != nil {
			return nil, err
		}
	}

	return &KV{
		medium: "memory",
		db:     memoryDB,
	}, nil
}

// GetDiskStore does what it says on the tin
func GetDiskStore() (*KV, error) {
	config := config.NewConfig()

	var err error
	if diskDB == nil {
		diskDB, err = badger.Open(badger.DefaultOptions(config.DBPath).WithSyncWrites(config.DBSyncWrites).WithLogger(&LogWrapper{}))
		if err != nil {
			return nil, err
		}
	}

	return &KV{
		medium: "disk",
		db:     diskDB,
	}, nil
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

// ListKeys up to the limit specified or error
func (kv *KV) ListKeys(limit int) ([]string, error) {
	results := make([]string, 0)

	count := 0
	err := kv.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			if count >= limit {
				break
			}

			results = append(results, string(it.Item().Key()))
			count++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ListPairs up to the limit specified or error
func (kv *KV) ListPairs(limit int) ([]Pair, error) {
	results := make([]Pair, 0)

	count := 0
	err := kv.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			if count >= limit {
				break
			}

			item := it.Item()
			value := ""
			err := it.Item().Value(func(val []byte) error {
				value = string(val)
				return nil
			})
			if err != nil {
				return err
			}

			results = append(results, Pair{
				Key:   string(item.Key()),
				Value: value,
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

// StartTransaction or error
func (kv *KV) StartTransaction() error {
	kv.transaction = kv.db.NewTransaction(true)

	return nil
}

// EndTransaction or error
func (kv *KV) EndTransaction() error {
	defer kv.transaction.Discard()

	err := kv.transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}

// StartSerialTransaction or error
func (kv *KV) StartSerialTransaction() error {
	if kv.medium == "disk" {
		diskSerialLock.Lock()
	} else if kv.medium == "memory" {
		memorySerialLock.Lock()
	}

	kv.transaction = kv.db.NewTransaction(true)
	return nil
}

// EndSerialTransaction or error
func (kv *KV) EndSerialTransaction() error {
	defer kv.transaction.Discard()
	if kv.medium == "disk" {
		defer diskSerialLock.Unlock()
	} else if kv.medium == "memory" {
		defer memorySerialLock.Unlock()
	}

	err := kv.transaction.Commit()
	if err != nil {
		return err
	}

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

// TransactionDelete the given key or error as part of a transaction
func (kv *KV) TransactionDelete(key string) error {
	err := kv.transaction.Delete([]byte(key))
	if err != nil {
		return err
	}

	return nil
}
