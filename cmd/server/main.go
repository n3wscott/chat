package main

import (
	"flag"
	"fmt"
	"github.com/n3wscott/chat/pkg/server"
)

var (
	port int
)

func main() {
	flag.Parse()
	done := make(chan bool, 1)
	s := server.NewServer("", port)

	go func() {
		if err := s.Run(); err != nil {
			fmt.Printf("Failed to listen: %v\n", err)
			done <- true
			return
		}
	}()

	fmt.Printf("Ready  :%d\n", port)
	<-done
}

func init() {
	flag.IntVar(&port, "port", 8080, "The http port")
}
