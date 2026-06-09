package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreSetGet(t *testing.T) {
	store := NewStore()

	store.Set("name", "ted")

	value, ok := store.Get("name")
	if !ok {
		t.Fatal("expected key to exits")
	}

	if value != "ted" {
		t.Fatalf("expected value %q, got %q", "ted", value)
	}
}

func TestStoreDel(t *testing.T) {
	store := NewStore()

	store.Set("name", "ted")

	if !store.Del("name") {
		t.Fatal("expected Del to return true")
	}

	if store.Del("name") {
		t.Fatal("expected Del to return false for missing key")
	}

	if store.Exists("name") {
		t.Fatal("expected key to be deleted")
	}
}

func TestStoreClearAndCount(t *testing.T) {
	store := NewStore()

	store.Set("name", "ted")
	store.Set("city", "berlin")

	if count := store.Count(); count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}

	store.Clear()

	if count := store.Count(); count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}

func TestStoreExpireAndTTL(t *testing.T) {
	store := NewStore()

	store.Set("session", "abc")

	if ttl := store.TTL("session"); ttl != -1 {
		t.Fatalf("expected TTL -1 before expiration, got %d", ttl)
	}

	if !store.Expire("session", 1) {
		t.Fatal("expected Expire to return true")
	}

	if ttl := store.TTL("session"); ttl < 0 || ttl > 1 {
		t.Fatalf("expected TTL between 0 and 1, got %d", ttl)
	}

	time.Sleep(1100 * time.Millisecond)

	value, ok := store.Get("session")
	if ok {
		t.Fatalf("expected expired key to be missing, got %q", value)
	}

	if ttl := store.TTL("session"); ttl != -2 {
		t.Fatalf("expected TTL -2 after expiration, got %d", ttl)
	}
}

func TestStoreDeleteExpired(t *testing.T) {
	store := NewStore()

	store.Set("temp", "hello")
	store.Expire("temp", 1)

	if deleted := store.DeleteExpired(); deleted != 0 {
		t.Fatalf("expected 0 deleted before expiration, got %d", deleted)
	}

	time.Sleep(1100 * time.Millisecond)

	if deleted := store.DeleteExpired(); deleted != 1 {
		t.Fatalf("expected 1 deleted after expiration, got %d", deleted)
	}

	if store.Exists("temp") {
		t.Fatal("expected expired key to be gone")
	}
}

func TestStoreSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dump.json")

	store := NewStore()
	store.Set("name", "ted")
	store.Set("city", "warsaw")

	if err := store.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded := NewStore()
	if err := loaded.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	value, ok := loaded.Get("name")
	if !ok {
		t.Fatal("expected loaded key to exist")
	}

	if value != "ted" {
		t.Fatalf("expected loaded value %q, got %q", "ted", value)
	}

	if count := loaded.Count(); count != 2 {
		t.Fatalf("expected loaded count 2, got %d", count)
	}
}
