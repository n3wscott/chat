package main

import (
	"bufio"
	"fmt"
	"github.com/n3wscott/chat/pkg/api"
	"github.com/n3wscott/chat/pkg/client"
	"log"
	"os"
)

func main() {
	c := client.NewClient("scott", "localhost", 1337)

	if err := c.Connect(); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	go func() {
		for {
			select {
			case m := <-c.Reader:
				log.Printf("%s: %s\n", m.Author, m.Body)
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		msg, _, err := reader.ReadLine()
		if err != nil {
			return
		}
		c.Writer <- api.Message{Author: "scott", Body: string(msg)}
	}
}
