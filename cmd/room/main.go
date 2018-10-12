package main

import (
	"flag"
	"github.com/n3wscott/chat/pkg/room"
)

var (
	host string
	port int
	name string
)

func main() {
	flag.Parse()

	r := room.NewRoom(name, host, port).Run()

	<-r.Done
}

func init() {
	flag.StringVar(&name, "name", "steve", "Your name.")
	flag.IntVar(&port, "port", 80, "The http port.")
	flag.StringVar(&host, "host", "localhost", "The http host.")
}
