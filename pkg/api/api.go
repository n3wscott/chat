package api

import (
	"encoding/json"
	"log"
)

type Message struct {
	Author string `json:"a"`
	Body   string `json:"b"`
}

func Parse(bytes []byte) Message {
	m := &Message{}
	if err := json.Unmarshal(bytes, m); err != nil {
		log.Printf("failed to parse: %v", err)
		return Message{}
	}
	return *m
}

func (m *Message) Json() []byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		log.Printf("failed to json: %v", err)
		return make([]byte, 0)
	}
	return bytes
}
