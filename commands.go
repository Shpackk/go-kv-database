package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const defaultDataFile = "dump.json"

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
			return encodeError("usage: SET key value")
		}

		key := parts[1]
		value := strings.Join(parts[2:], " ")

		h.store.Set(key, value)
		return encodeSimpleString("OK")

	case "GET":
		if len(parts) != 2 {
			return encodeError("usage: GET key")
		}

		key := parts[1]

		value, ok := h.store.Get(key)
		if !ok {
			return encodeNil()
		}

		return encodeBulkString(value)

	case "DEL":
		if len(parts) != 2 {
			return encodeError("usage: DEL key")
		}

		key := parts[1]

		deleted := h.store.Del(key)
		if !deleted {
			return encodeInteger(0)
		}

		return encodeInteger(1)

	case "EXISTS":
		if len(parts) != 2 {
			return encodeError("usage: EXISTS key")
		}

		key := parts[1]

		if h.store.Exists(key) {
			return encodeInteger(1)
		}

		return encodeInteger(0)

	case "KEYS":
		if len(parts) != 1 {
			return encodeError("usage: KEYS")
		}

		keys := h.store.Keys()
		if len(keys) == 0 {
			return encodeArray([]string{})
		}

		return encodeArray(keys)

	case "COUNT":
		if len(parts) != 1 {
			return encodeError("usage: COUNT")
		}

		return encodeInteger(h.store.Count())

	case "CLEAR":
		if len(parts) != 1 {
			return encodeError("usage: CLEAR")
		}

		h.store.Clear()
		return encodeSimpleString("OK")

	case "EXPIRE":
		if len(parts) != 3 {
			return encodeError("usage: EXPIRE key seconds")
		}

		key := parts[1]

		seconds, err := strconv.Atoi(parts[2])
		if err != nil {
			return encodeError("seconds must be an int")
		}

		if seconds <= 0 {
			return encodeError("seconds must be greater than 0")
		}

		ok := h.store.Expire(key, seconds)
		if !ok {
			return encodeInteger(0)
		}

		return encodeInteger(1)

	case "TTL":
		if len(parts) != 2 {
			return encodeError("usage: TTL key")
		}

		key := parts[1]

		return encodeInteger(h.store.TTL(key))

	case "SAVE":
		if len(parts) > 2 {
			return encodeError("usage: SAVE [path]")
		}

		path := defaultDataFile
		if len(path) == 2 {
			path = parts[1]
		}

		if err := h.store.Save(path); err != nil {
			return encodeError(err.Error())
		}

		return encodeSimpleString("OK")

	case "LOAD":
		if len(parts) > 2 {
			return encodeError("usage: LOAD [path]")
		}

		path := defaultDataFile
		if len(parts) == 2 {
			path = parts[1]
		}

		if err := h.store.Load(path); err != nil {
			return encodeError(err.Error())
		}

		return encodeSimpleString("OK")

	case "PING":
		return encodeSimpleString("PONG")

	case "EXIT":
		return encodeSimpleString("OK")

	default:
		return encodeError("unknown command")
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

func encodeSimpleString(value string) string {
	return fmt.Sprintf("+%s\r\n", value)
}

func encodeError(message string) string {
	return fmt.Sprintf("-ERR %s\r\n", message)
}

func encodeInteger(value int) string {
	return fmt.Sprintf(":%d\r\n", value)
}

func encodeBulkString(value string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
}

func encodeNil() string {
	return "$-1\r\n"
}

func encodeArray(values []string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("*%d\r\n", len(values)))

	for _, value := range values {
		builder.WriteString((encodeBulkString(value)))
	}

	return builder.String()
}
