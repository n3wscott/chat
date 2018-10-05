package main

import (
	"fmt"
	"github.com/n3wscott/chat/pkg/server"
)

func main() {
	done := make(chan bool, 1)
	s := server.NewServer("localhost", 1337)
	if err := s.Listen(); err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		return
	}
	<-done
}
