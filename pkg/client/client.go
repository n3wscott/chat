package client

import (
	"bytes"
	"fmt"
	"github.com/n3wscott/chat/pkg/api"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func NewClient(name, host string, port int) *Client {
	var url string
	if port == 80 {
		url = fmt.Sprintf("http://%s", host)
	} else if port == 443 {
		url = fmt.Sprintf("https://%s", host)
	} else {
		url = fmt.Sprintf("http://%s:%d", host, port)
	}
	return &Client{
		name: name,
		url:  url,
		Here: make(chan string, 10),
		Msg:  make(chan api.Message, 10),
		Tx:   make(chan api.Message, 10),
		Done: make(chan bool),
	}
}

type Client struct {
	name string
	url  string

	Here chan string
	Msg  chan api.Message
	Tx   chan api.Message
	Done chan bool
}

func (c *Client) Run() {
	readDone := make(chan bool)
	writeDone := make(chan bool)

	go c.doReader(readDone)
	go c.doWriter(writeDone)

	select {
	case <-c.Done:
		readDone <- true
		writeDone <- true
	}
}

func (c *Client) doReader(done chan bool) {

	for {
		// Check done was called in a non-blocking way.
		select {
		case <-done:
			return
		default:
			// continue
		}

		url := fmt.Sprintf("%s/chat?me=%s&wait=%s", c.url, c.name, 30*time.Second)
		res, err := http.Get(url)
		if err != nil {
			log.Print(err)
			continue
		}
		resourceResp, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Print(err)
			continue
		}

		if res.StatusCode == 200 {
			room := api.Parse(resourceResp)
			for _, h := range room.Here {
				c.Here <- h
			}
			for _, m := range room.Messages {
				c.Msg <- m
			}
		}
	}

	return
}

func (c *Client) doWriter(done chan bool) {
	for {
		select {
		case m := <-c.Tx:
			_, err := http.Post(c.url+"/msg", "application/json", bytes.NewBuffer(m.Json()))
			if err != nil {
				log.Print(err)
				continue
			}
		case <-done:
			return
		}
	}
}
