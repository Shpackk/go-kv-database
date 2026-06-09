package main

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestHandler(t *testing.T) *CommandHandler {
	t.Helper()

	dir := t.TempDir()
	dataPath := filepath.Join(dir, "dump.json")
	aofPath := filepath.Join(dir, "appendonly.aof")

	return NewCommandHandler(NewStore(), dataPath, aofPath)
}

func TestCommandSetGet(t *testing.T) {
	handler := newTestHandler(t)

	if response := handler.Handle("SET name ted"); response != "+OK\r\n" {
		t.Fatalf("expected SET response OK, got %q", response)
	}

	if response := handler.Handle("GET name"); response != "$3\r\nted\r\n" {
		t.Fatalf("expected GET bulk string, got %q", response)
	}
}

func TestCommandQuotedString(t *testing.T) {
	handler := newTestHandler(t)

	if response := handler.Handle(`SET message "hello world"`); response != "+OK\r\n" {
		t.Fatalf("expected SET response OK, got %q", response)
	}

	if response := handler.Handle("GET message"); response != "$11\r\nhello world\r\n" {
		t.Fatalf("expected quoted value to round-trip, got %q", response)
	}
}

func TestCommandMissingKey(t *testing.T) {
	handler := newTestHandler(t)

	if response := handler.Handle("GET missing"); response != "$-1\r\n" {
		t.Fatalf("expected nil response, got %q", response)
	}
}

func TestCommandExpire(t *testing.T) {
	handler := newTestHandler(t)

	handler.Handle("SET session abc")

	if response := handler.Handle("EXPIRE session 10"); response != ":1\r\n" {
		t.Fatalf("expected successful expire response, got %q", response)
	}

	if response := handler.Handle("TTL session"); response == ":-1\r\n" || response == ":-2\r\n" {
		t.Fatalf("expected active TTL, got %q", response)
	}
}

func TestCommandParserRejectsUnterminatedQuote(t *testing.T) {
	_, err := parseCommandLine(`SET name "hello`)
	if err == nil {
		t.Fatal("expected unterminated quote error")
	}
}

func TestCommandSaveLoad(t *testing.T) {
	handler := newTestHandler(t)

	handler.Handle("SET name ted")

	if response := handler.Handle("SAVE"); response != "+OK\r\n" {
		t.Fatalf("expected SAVE response OK, got %q", response)
	}

	handler.Handle("CLEAR")

	if response := handler.Handle("GET name"); response != "$-1\r\n" {
		t.Fatalf("expected key to be gone after CLEAR, got %q", response)
	}

	if response := handler.Handle("LOAD"); response != "+OK\r\n" {
		t.Fatalf("expected LOAD response OK, got %q", response)
	}

	if response := handler.Handle("GET name"); response != "$3\r\nted\r\n" {
		t.Fatalf("expected loaded value, got %q", response)
	}
}

func TestCommandAOFReplay(t *testing.T) {
	handler := newTestHandler(t)

	if response := handler.Handle("AOFON"); response != "+OK\r\n" {
		t.Fatalf("expected AOFON response OK, got %q", response)
	}

	handler.Handle("SET name ted")
	handler.Handle("SET city warsaw")
	handler.Handle("AOFOFF")
	handler.Handle("CLEAR")

	if response := handler.Handle("LOADAOF"); response != "+OK\r\n" {
		t.Fatalf("expected LOADAOF response OK, got %q", response)
	}

	if response := handler.Handle("GET name"); response != "$3\r\nted\r\n" {
		t.Fatalf("expected name from AOF, got %q", response)
	}

	if response := handler.Handle("GET city"); response != "$6\r\nwarsaw\r\n" {
		t.Fatalf("expected city from AOF, got %q", response)
	}
}

func TestCommandAOFReplayWithCustomPath(t *testing.T) {
	dir := t.TempDir()
	defaultDataPath := filepath.Join(dir, "dump.json")
	defaultAOFPath := filepath.Join(dir, "default.aof")
	customAOFPath := filepath.Join(dir, "custom.aof")

	handler := NewCommandHandler(NewStore(), defaultDataPath, defaultAOFPath)

	if response := handler.Handle("AOFON " + customAOFPath); response != "+OK\r\n" {
		t.Fatalf("expected AOFON response OK, got %q", response)
	}

	handler.Handle("SET name ted")

	if _, err := os.Stat(customAOFPath); err != nil {
		t.Fatalf("expected custom AOF file to exist: %v", err)
	}

	handler.Handle("AOFOFF")
	handler.Handle("CLEAR")

	if response := handler.Handle("LOADAOF " + customAOFPath); response != "+OK\r\n" {
		t.Fatalf("expected custom LOADAOF response OK, got %q", response)
	}

	if response := handler.Handle("GET name"); response != "$3\r\nted\r\n" {
		t.Fatalf("expected value from custom AOF path, got %q", response)
	}
}
