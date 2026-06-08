package main

func main() {
	store := NewStore()
	handler := NewCommandHandler(store)

	server := NewServer(":6379", handler)
	server.Start()
}
