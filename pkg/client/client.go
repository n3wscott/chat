package client

import (
	"fmt"
	"log"
	"net"
)

func NewClient(name, host string, port int) *Client {
	return &Client{
		name:    name,
		network: "tcp",
		address: fmt.Sprintf("%s:%d", host, port),
		Reader:  make(chan []byte),
		Writer:  make(chan []byte),
		Done:    make(chan bool),
	}
}

type Client struct {
	name string

	network    string
	address    string
	connection net.Conn

	listener net.Listener

	Reader chan []byte
	Writer chan []byte
	Done   chan bool
}

func (c *Client) Connect() error {
	conn, err := net.Dial(c.network, c.address)
	if err != nil {
		return err
	}
	c.connection = conn

	go c.onConnect()
	return nil
}

func (c *Client) onConnect() {
	defer c.connection.Close()

	_, err := c.connection.Write([]byte(c.name))
	if err != nil {
		fmt.Printf("Error in connection: %v\n", err)
		return
	}

	go c.doRead()
	for {
		select {
		case r := <-c.Reader:
			log.Println(string(r))
		case w := <-c.Writer:
			_, err := c.connection.Write(w)
			if err != nil {
				fmt.Printf("Error writing to connection: %v\n", err)
				c.Done <- true
			}
		case <-c.Done:
			return
		}
	}
}

func (c *Client) doRead() {
	msg := make([]byte, 1024)
	for {
		length, err := c.connection.Read(msg) // TODO: make this an object.
		if err != nil {
			fmt.Printf("Error reading from connection: %v\n", err)
			c.Done <- true
			return
		}
		c.Reader <- msg[:length]
	}
}
