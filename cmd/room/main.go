package room

import (
	"github.com/n3wscott/chat/pkg/api"
	"github.com/n3wscott/chat/pkg/room"
)

func main() {
	done := make(chan bool, 1)

	r := room.NewRoom(done, "scott").Run()

	r.Room <- api.Message{Body: "Welcome to the room!"}

	go func() {
		for {
			select {
			case msg := <-r.Entry:
				r.Room <- api.Message{Author: "scott", Body: msg.Body}
			}
		}
	}()

	<-done
}
