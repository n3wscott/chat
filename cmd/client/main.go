package main

import (
	"bufio"
	"flag"
	"github.com/n3wscott/chat/pkg/api"
	"github.com/n3wscott/chat/pkg/client"
	"log"
	"os"
)

var (
	host string
	port int
	name string
)

func main() {
	flag.Parse()

	c := client.NewClient(name, host, port)

	go c.Run()
	defer func() {
		c.Done <- true
	}()

	go func() {
		here := []string(nil)
		for {
			select {
			case m := <-c.Msg:
				log.Printf("%s: %s\n", m.Author, m.Body)
			case h := <-c.Here:
				var n bool
				here, n = addNew(here, h)
				if n {
					log.Printf("%s just joined the room.\n", h)
				}
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		msg, _, err := reader.ReadLine()
		if err != nil {
			continue
		}
		c.Tx <- api.Message{Author: name, Body: string(msg)}
	}
}

func addNew(all []string, new string) ([]string, bool) {
	for _, a := range all {
		if a == new {
			return all, false
		}
	}
	all = append(all, new)
	return all, true
}

func init() {
	flag.StringVar(&name, "name", "scott", "Your name.")
	flag.IntVar(&port, "port", 8080, "The http port.")
	flag.StringVar(&host, "host", "localhost", "The http host.")
}
