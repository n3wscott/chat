package main

import (
	"bufio"
	"fmt"
	"github.com/n3wscott/chat/pkg/client"
	"os"
)

func main() {
	c := client.NewClient("scott", "localhost", 1337)

	if err := c.Connect(); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		msg, _, err := reader.ReadLine()
		if err != nil {
			return
		}
		c.Writer <- msg
	}
}
