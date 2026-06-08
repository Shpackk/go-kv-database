package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	store := make(map[string]string)

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("kv-db started. Type EXIT to quit.")

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.ToUpper(line) == "EXIT" {
			fmt.Println("Exiting.")
			break
		}

		parts := strings.Fields(line)
		command := strings.ToUpper(parts[0])

		switch command {
		case "SET":
			if len(parts) < 3 {
				fmt.Println("ERR! Usage: SET key value")
				continue
			}

			key := parts[1]
			value := strings.Join(parts[2:], " ")

			store[key] = value

			fmt.Println("OK")

		case "GET":
			if len(parts) != 2 {
				fmt.Println("ERR! Usage: GET key")
				continue
			}

			key := parts[1]

			value, ok := store[key]
			if !ok {
				fmt.Println("(nil)")
			}

			fmt.Println(value)

		case "DEL":
			if len(parts) != 2 {
				fmt.Println("ERR! Usage: DEL key")
				continue
			}

			key := parts[1]

			_, ok := store[key]
			if !ok {
				fmt.Println("0")
				continue
			}

			delete(store, key)
			fmt.Println("1")

		case "EXISTS":
			if len(parts) != 2 {
				fmt.Println("ERR! Usage: EXISTS key")
				continue
			}

			key := parts[1]

			_, ok := store[key]

			if !ok {
				fmt.Println("0")
			} else {
				fmt.Println("1")
			}

		case "KEYS":
			if len(parts) != 1 {
				fmt.Println("ERR! Usage: KEYS")
				continue
			}

			if len(store) == 0 {
				fmt.Println("(empty)")
				continue
			}

			for key := range store {
				fmt.Println(key)
			}

		case "COUNT":
			if len(parts) != 1 {
				fmt.Println("ERR! Usage: COUNT")
				continue
			}

			count := len(store)
			fmt.Println(count)

		case "CLEAR":
			if len(parts) != 1 {
				fmt.Println("ERR! Usage: CLEAR")
				continue
			}

			store = make(map[string]string)
			fmt.Println("OK")

		default:
			fmt.Println("ERR! Unknown command")
		}

		_ = store
	}
}
