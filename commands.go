package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type CommandHandler struct {
	store *Store
}

func NewCommandHandler(store *Store) *CommandHandler {
	return &CommandHandler{
		store: store,
	}
}

func (h *CommandHandler) Handle(line string) string {
	parts, err := parseCommandLine(line)
	if err != nil {
		return "ERR! " + err.Error()
	}

	if len(parts) == 0 {
		return ""
	}

	command := strings.ToUpper(parts[0])

	switch command {
	case "SET":
		if len(parts) < 3 {
			return "ERR! usage: SET key value"
		}

		key := parts[1]
		value := strings.Join(parts[2:], " ")

		h.store.Set(key, value)
		return "OK"

	case "GET":
		if len(parts) != 2 {
			return "ERR! usage: GET key"
		}

		key := parts[1]

		value, ok := h.store.Get(key)
		if !ok {
			return "(nil)"
		}

		return value

	case "DEL":
		if len(parts) != 2 {
			return "ERR! usage: DEL key"
		}

		key := parts[1]

		deleted := h.store.Del(key)
		if !deleted {
			return "0"
		}

		return "1"

	case "EXISTS":
		if len(parts) != 2 {
			return "ERR! usage: EXISTS key"
		}

		key := parts[1]

		if h.store.Exists(key) {
			return "1"
		}

		return "0"

	case "KEYS":
		if len(parts) != 1 {
			return "ERR! usage: KEYS"
		}

		keys := h.store.Keys()
		if len(keys) == 0 {
			return "(empty)"
		}

		return strings.Join(keys, " ")

	case "COUNT":
		if len(parts) != 1 {
			return "ERR! usage: COUNT"
		}

		return fmt.Sprintf("%d", h.store.Count())

	case "CLEAR":
		if len(parts) != 1 {
			return "ERR! usage: CLEAR"
		}

		h.store.Clear()
		return "OK"

	case "EXPIRE":
		if len(parts) != 3 {
			return "ERR! usage: EXPIRE key seconds"
		}

		key := parts[1]

		seconds, err := strconv.Atoi(parts[2])
		if err != nil {
			return "ERR! seconds must be an int"
		}

		if seconds <= 0 {
			return "ERR! seconds must be greater than 0"
		}

		ok := h.store.Expire(key, seconds)
		if !ok {
			return "0"
		}

		return "1"

	case "TTL":
		if len(parts) != 2 {
			return "ERR! usage: TTL key"
		}

		key := parts[1]

		return fmt.Sprintf("%d", h.store.TTL(key))

	case "PING":
		return "PONG"

	case "EXIT":
		return "dont come back"

	default:
		return "ERR! unknown command"
	}
}

func parseCommandLine(line string) ([]string, error) {
	var parts []string
	var current strings.Builder

	inQuotes := false
	tokenStarted := false

	for _, ch := range line {
		if ch == '"' {
			inQuotes = !inQuotes
			tokenStarted = true
			continue
		}

		if unicode.IsSpace(ch) && !inQuotes {
			if tokenStarted {
				parts = append(parts, current.String())
				current.Reset()
				tokenStarted = false
			}
			continue
		}

		current.WriteRune(ch)
		tokenStarted = true
	}

	if inQuotes {
		return nil, fmt.Errorf("unterminated quote")
	}

	if tokenStarted {
		parts = append(parts, current.String())
	}

	return parts, nil
}
