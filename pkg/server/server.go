package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/n3wscott/chat/pkg/api"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func NewServer(host string, port int) *Server {
	return &Server{
		address: fmt.Sprintf("%s:%d", host, port),
		clients: make(map[string]chan bool, 5),
		seen:    make([]string, 0),
	}
}

type Server struct {
	address string

	clients map[string]chan bool

	server *http.Server

	message *api.Message // TODO: this should be a ring buffer.
	seen    []string
}

func (s *Server) Run() error {

	http.HandleFunc("/chat", s.roomFunc)
	http.HandleFunc("/msg", s.chatFunc)

	s.server = &http.Server{
		Addr:         s.address,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	return s.server.ListenAndServe()
}

func (s *Server) Close() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *Server) getLongPollDuration(r *http.Request) time.Duration {
	timeout, err := time.ParseDuration(r.URL.Query().Get("wait"))
	if err != nil {
		return 55 * time.Second
	}

	log.Printf("found custom timeout: %s", timeout)
	return timeout
}

func (s *Server) addClient(id string) chan bool {
	if _, ok := s.clients[id]; !ok {
		s.clients[id] = make(chan bool)
	}
	return s.clients[id]
}

func (s *Server) removeClient(id string) {
	delete(s.clients, id)
}

func (s *Server) waitForMessages(ctx context.Context, wait time.Duration) []api.Message {
	timeout := time.Tick(wait)

	clientId := randString(8)
	update := s.addClient(clientId)
	defer s.removeClient(clientId)

	select {
	case <-ctx.Done():
		log.Printf("context cancel")
		return []api.Message(nil)
	case <-timeout:
		log.Printf("method timeout: %s", clientId)
		return []api.Message(nil)
	case <-update:
		log.Printf("update for: %s", clientId)
		return []api.Message{*s.message}
	}
}

func (s *Server) chatFunc(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Print(err)
		return
	}
	m := &api.Message{}
	if err := json.Unmarshal(body, m); err != nil {
		log.Print(err)
		return
	}
	s.broadcast(m)
}

func (s *Server) addHere(h string) {
	for _, s := range s.seen {
		if s == h {
			return
		}
	}
	s.seen = append(s.seen, h) // TODO: would be cool to add a timeout
}

func (s *Server) roomFunc(w http.ResponseWriter, r *http.Request) {
	timeout := s.getLongPollDuration(r)
	me := r.URL.Query().Get("me")

	s.addHere(me)

	messages := s.waitForMessages(r.Context(), timeout)
	if len(messages) == 0 {
		// write long poll timeout
		w.WriteHeader(http.StatusNotModified)
	}
	w.Header().Set("Content-Type", "application/json")
	// do not cache response
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	room := &api.Room{
		Here:     s.seen,
		Messages: messages,
	}

	js, err := json.Marshal(room)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (s *Server) broadcast(m *api.Message) {
	s.message = m

	for _, c := range s.clients {
		c <- true
	}
}
