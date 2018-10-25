package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/knative/eventing/pkg/event"
	"github.com/n3wscott/chat/pkg/api"
	"github.com/n3wscott/chat/pkg/client"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	sink string
	host string
	port int
	name string
)

func init() {
	flag.StringVar(&name, "name", "bot", "Your name.")
	flag.StringVar(&sink, "sink", "", "the host url to heartbeat to")
	flag.StringVar(&host, "host", "localhost", "The http host of the chat server.")
	flag.IntVar(&port, "port", 8080, "The http port.")
}

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
				postMessage(sink, "botless.chat.msg", &m)
			case h := <-c.Here:
				var n bool
				here, n = addNew(here, h)
				if n {
					postMessage(sink, "botless.chat.join", &api.Message{Author: h})
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

// Creates a CloudEvent Context for a given heartbeat.
func cloudEventsContext(eventType string, m *api.Message) *event.EventContext {
	return &event.EventContext{
		CloudEventsVersion: event.CloudEventsVersion,
		EventType:          eventType,
		EventID:            uuid.New().String(),
		Source:             fmt.Sprintf("%s@%s", name, host),
		EventTime:          time.Now(),
	}
}

func postMessage(target string, eventType string, m *api.Message) error {
	ctx := cloudEventsContext(eventType, m)

	log.Printf("posting to %q", target)
	req, err := event.Binary.NewRequest(target, m, *ctx)
	if err != nil {
		log.Printf("failed to create http request: %s", err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to do POST: %v", err)
		return err
	}
	defer resp.Body.Close()
	log.Printf("response Status: %s", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("response Body: %s", string(body))
	return nil
}
