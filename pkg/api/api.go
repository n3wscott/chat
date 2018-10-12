package api

import (
	"encoding/json"
	"log"
)

type Room struct {
	Here     []string  `json:"here"`
	Messages []Message `json:"msg"`
}

type Message struct {
	ID     string `json:"id"`
	Author string `json:"author"`
	Body   string `json:"body"`
}

func Parse(bytes []byte) *Room {
	r := &Room{}
	if err := json.Unmarshal(bytes, &r); err != nil {
		log.Printf("failed to parse %s: %v", string(bytes), err)
		return &Room{}
	}
	return r
}

func (m *Message) Json() []byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		log.Printf("failed to json: %v", err)
		return make([]byte, 0)
	}
	return bytes
}
