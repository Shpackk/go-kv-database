package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type CommandHandler struct {
	store    *Store
	dataPath string
	aofPath  string // appendOnlyFile (log)
	aofOn    bool   // appendOnlyFile (log) turn off/on
}

func NewCommandHandler(store *Store, dataPath string, aofPath string) *CommandHandler {
	return &CommandHandler{
		store:    store,
		dataPath: dataPath,
		aofPath:  aofPath,
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
		if err := h.appendAOF(line); err != nil {
			return encodeError(err.Error())
		}
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

		if err := h.appendAOF(line); err != nil {
			return encodeError(err.Error())
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
		if err := h.appendAOF(line); err != nil {
			return encodeError(err.Error())
		}
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

		if err := h.appendAOF(line); err != nil {
			return encodeError(err.Error())
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

		path := h.dataPath
		if len(parts) == 2 {
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

		path := h.dataPath
		if len(parts) == 2 {
			path = parts[1]
		}

		if err := h.store.Load(path); err != nil {
			return encodeError(err.Error())
		}

		return encodeSimpleString("OK")

	case "AOFON":
		if len(parts) > 2 {
			return encodeError("usage: AOFON [path]")
		}

		if len(parts) == 2 {
			h.aofPath = parts[1]
		}

		h.aofOn = true
		return encodeSimpleString("OK")

	case "AOFOFF":
		if len(parts) != 1 {
			return encodeError("usage: AOFOFF")
		}

		h.aofOn = false
		return encodeSimpleString("OK")

	case "LOADAOF":
		if len(parts) > 2 {
			return encodeError("usage: LOADAOF [path]")
		}

		path := h.aofPath
		if len(parts) == 2 {
			path = parts[1]
		}

		if err := h.loadAOF(path); err != nil {
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
	escaping := false
	tokenStarted := false

	for _, ch := range line {
		if escaping {
			switch ch {
			case 'n':
				current.WriteRune('\n')
			case 'r':
				current.WriteRune('\r')
			case 't':
				current.WriteRune('\t')
			case '"':
				current.WriteRune('"')
			case '\\':
				current.WriteRune('\\')
			default:
				current.WriteRune(ch)
			}
			escaping = false
			tokenStarted = true
			continue
		}

		if inQuotes && ch == '\\' {
			escaping = true
			tokenStarted = true
			continue
		}

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

	if escaping {
		current.WriteRune('\\')
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

func (h *CommandHandler) appendAOF(line string) error {
	if !h.aofOn {
		return nil
	}

	file, err := os.OpenFile(h.aofPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(line + "\n")
	return err
}

func (h *CommandHandler) loadAOF(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	wasAOFOn := h.aofOn
	h.aofOn = false
	defer func() {
		h.aofOn = wasAOFOn
	}()

	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		response := h.Handle(line)
		if strings.HasPrefix(response, "-ERR") {
			return fmt.Errorf("failed to replay %q: %s", line, strings.TrimSpace(response))
		}
	}

	return nil
}
