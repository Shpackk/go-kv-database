package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestReadCommandInline(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("GET name\r\n"))

	line, err := readCommand(reader)
	if err != nil {
		t.Fatalf("readCommand failed: %v", err)
	}

	if line != "GET name" {
		t.Fatalf("expected inline command, got %q", line)
	}
}

func TestReadCommandRESPArray(t *testing.T) {
	input := "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	line, err := readCommand(reader)
	if err != nil {
		t.Fatalf("readCommand failed: %v", err)
	}

	if line != "GET name" {
		t.Fatalf("expected RESP command to become inline command, got %q", line)
	}
}

func TestReadCommandRESPArrayWithSpaces(t *testing.T) {
	input := "*3\r\n$3\r\nSET\r\n$7\r\nmessage\r\n$11\r\nhello world\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	line, err := readCommand(reader)
	if err != nil {
		t.Fatalf("readCommand failed: %v", err)
	}

	if line != `SET message "hello world"` {
		t.Fatalf("expected quoted inline command, got %q", line)
	}
}

func TestReadCommandRESPArrayWithQuote(t *testing.T) {
	input := "*3\r\n$3\r\nSET\r\n$5\r\nquote\r\n$11\r\nhello \"ted\"\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	line, err := readCommand(reader)
	if err != nil {
		t.Fatalf("readCommand failed: %v", err)
	}

	parts, err := parseCommandLine(line)
	if err != nil {
		t.Fatalf("parseCommandLine failed: %v", err)
	}

	if parts[2] != `hello "ted"` {
		t.Fatalf("expected quoted value to round-trip, got %q", parts[2])
	}
}

func TestReadCommandRejectsBadBulkTerminator(t *testing.T) {
	input := "*1\r\n$3\r\nGETxx"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := readCommand(reader)
	if err == nil {
		t.Fatal("expected bad bulk terminator error")
	}
}
