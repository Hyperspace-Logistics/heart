package kv_test

import (
	"testing"

	"github.com/sosodev/heart/kv"
)

func TestKV(t *testing.T) {
	kv, err := kv.GetMemoryStore()
	if err != nil {
		t.Fatalf("failed to get memory store: %s", err)
	}

	value, err := kv.Get("test-key")
	if err != nil {
		t.Fatalf("failed to get key: %s", err)
	}

	if value != "" {
		t.Error("initial value should be empty")
	}

	err = kv.StartTransaction()
	if err != nil {
		t.Fatalf("failed to start transaction: %s", err)
	}

	err = kv.TransactionSet("test-key", "Hello, world!")
	if err != nil {
		t.Fatalf("failed to set key: %s", err)
	}

	value, err = kv.TransactionGet("test-key")
	if err != nil {
		t.Fatalf("failed to get key: %s", err)
	}

	if value != "Hello, world!" {
		t.Error("incorrect transaction value after set")
	}

	err = kv.EndTransaction()
	if err != nil {
		t.Fatalf("failed to end transaction: %s", err)
	}

	value, err = kv.Get("test-key")
	if err != nil {
		t.Fatalf("failed to get key after transaction: %s", err)
	}

	if value != "Hello, world!" {
		t.Errorf("incorrect read value outside of transaction, expected %s got %s", "Hello, world!", value)
	}
}
