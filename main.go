package main

import "time"

func main() {
	store := NewStore()
	stopExpiration := store.StartActiveExpiration(1 * time.Second)
	defer stopExpiration()

	handler := NewCommandHandler(store)

	server := NewServer(":6379", handler)
	server.Start()
}
