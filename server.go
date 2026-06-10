package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Server struct {
	addr    string
	handler *CommandHandler
}

func NewServer(addr string, handler *CommandHandler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		fmt.Println("ERR! Failed to start server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("kv-db server listening on", s.addr)
	fmt.Println("connect with: ncat localhost", strings.TrimPrefix(s.addr, ":"))

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("ERR! Faield to accept connection", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	addr := conn.RemoteAddr().String()

	fmt.Println("client connected:", addr)

	defer func() {
		conn.Close()
		fmt.Println("client disconnected", addr)
	}()

	reader := bufio.NewReader(conn)

	for {
		line, err := readCommand(reader)
		if err != nil {
			fmt.Println("ERR! connection read failed from", addr+":", err)
			return
		}

		if line == "" {
			continue
		}

		response := s.handler.Handle(line)

		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("ERR! failed to write response to", addr+":", err)
			return
		}

		if strings.ToUpper(line) == "EXIT" {
			return
		}
	}
}
