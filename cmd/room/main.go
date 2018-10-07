package main

import (
	"github.com/n3wscott/chat/pkg/room"
)

func main() {
	name := "steve"

	r := room.NewRoom(name, "localhost", 1337).Run()

	<-r.Done
}
