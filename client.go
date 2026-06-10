package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func RunClient(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	defer conn.Close()

	fmt.Println("connected to", addr)
	fmt.Println("type EXIT to quit")

	input := bufio.NewScanner(os.Stdin)
	reader := bufio.NewReader(conn)

	for {
		fmt.Print("kv> ")

		if !input.Scan() {
			return input.Err()
		}

		line := strings.TrimSpace(input.Text())
		if line == "" {
			continue
		}

		if _, err := fmt.Fprintln(conn, line); err != nil {
			return err
		}

		response, err := readClientResponse(reader)
		if err != nil {
			return err
		}

		fmt.Print(response)

		if strings.EqualFold(line, "EXIT") {
			return nil
		}
	}
}

func readClientResponse(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return "\n", nil
	}

	switch line[0] {
	case '+':
		return line[1:] + "\n", nil

	case '-':
		return line + "\n", nil

	case ':':
		return line[1:] + "\n", nil

	case '$':
		return readClientBulkString(reader, line)

	case '*':
		return readClientArray(reader, line)

	default:
		return line + "\n", nil
	}
}

func readClientBulkString(reader *bufio.Reader, header string) (string, error) {
	size, err := strconv.Atoi(strings.TrimPrefix(header, "$"))
	if err != nil {
		return "", err
	}

	if size == -1 {
		return "(nil)\n", nil
	}

	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimRight(value, "\r\n") + "\n", nil
}

func readClientArray(reader *bufio.Reader, header string) (string, error) {
	count, err := strconv.Atoi(strings.TrimPrefix(header, "*"))
	if err != nil {
		return "", err
	}

	var builder strings.Builder

	for i := 0; i < count; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimRight(line, "\r\n")

		if strings.HasPrefix(line, "$") {
			value, err := readClientBulkString(reader, line)
			if err != nil {
				return "", err
			}

			builder.WriteString(value)
			continue
		}

		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
