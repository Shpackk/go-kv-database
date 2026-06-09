package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	port := flag.Int("port", 6379, "server port")
	dataDir := flag.String("data", ".", "data directory")
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		fmt.Println("ERR! failed to create data directory", err)
		return
	}

	dataPath := filepath.Join(*dataDir, "dump.json")
	aofPath := filepath.Join(*dataDir, "appendonly.aof")
	addr := fmt.Sprintf(":%d", *port)

	store := NewStore()
	stopExpiration := store.StartActiveExpiration(1 * time.Second)
	defer stopExpiration()

	handler := NewCommandHandler(store, dataPath, aofPath)

	server := NewServer(addr, handler)
	server.Start()
}
