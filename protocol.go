package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

func readCommand(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return "", nil
	}

	if !strings.HasPrefix(line, "*") {
		return line, nil
	}

	parts, err := readRESPArray(reader, line)
	if err != nil {
		return "", err
	}

	return quoteCommandLine(parts), nil
}

func readRESPArray(reader *bufio.Reader, header string) ([]string, error) {
	count, err := strconv.Atoi(strings.TrimPrefix(header, "*"))
	if err != nil {
		return nil, fmt.Errorf("invalid RESP array header")
	}

	if count < 0 {
		return nil, fmt.Errorf("RESP arrays with negative length are not supported")
	}

	parts := make([]string, 0, count)

	for i := 0; i < count; i++ {
		bulkHeader, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		bulkHeader = strings.TrimRight(bulkHeader, "\r\n")
		if !strings.HasPrefix(bulkHeader, "$") {
			return nil, fmt.Errorf("expected bulk string")
		}

		size, err := strconv.Atoi(strings.TrimPrefix(bulkHeader, "$"))
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string size")
		}

		if size < 0 {
			return nil, fmt.Errorf("RESP bulk strings with negative length are not supported")
		}

		value := make([]byte, size+2)
		if _, err := io.ReadFull(reader, value); err != nil {
			return nil, err
		}

		if string(value[size:]) != "\r\n" {
			return nil, fmt.Errorf("invalid bulk string terminator")
		}

		parts = append(parts, string(value[:size]))
	}

	return parts, nil
}

func quoteCommandLine(parts []string) string {
	quoted := make([]string, 0, len(parts))

	for _, part := range parts {
		if needsQuoting(part) {
			quoted = append(quoted, strconv.Quote(part))
			continue
		}

		quoted = append(quoted, part)
	}

	return strings.Join(quoted, " ")
}

func needsQuoting(value string) bool {
	if value == "" {
		return true
	}

	for _, ch := range value {
		if unicode.IsSpace(ch) || ch == '"' {
			return true
		}
	}

	return false
}
