package server

import (
	"fmt"
	"net"
)

func NewServer(host string, port int) *Server {
	return &Server{
		network: "tcp",
		address: fmt.Sprintf("%s:%d", host, port),
	}
}

type Server struct {
	network     string
	address     string
	connections []net.Conn

	listener net.Listener
}

func (s *Server) Listen() error {

	if l, err := net.Listen("tcp", s.address); err != nil {
		return err
	} else {
		s.listener = l
		go s.listen()
	}
	return nil
}

func (s *Server) Close() {
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client: ", err.Error())
		}
		s.connections = append(s.connections, conn) // TODO: this should be a map.
		go s.onAccept(conn)
	}
}

func (s *Server) onAccept(conn net.Conn) {
	data := make([]byte, 1024) // TODO: for large messages this will fail.
	fmt.Println("Connected: ", conn.RemoteAddr())

	length, err := conn.Read(data)
	if err != nil {
		s.onClose(conn, "")
		return
	}
	name := string(data[:length])
	s.broadcast(conn, name+" has connected")

	for {
		length, err := conn.Read(data)
		if err != nil {
			s.onClose(conn, name)
			return
		}
		s.say(conn, name, string(data[:length]))
	}
}

func (s *Server) onClose(conn net.Conn, name string) {
	fmt.Println("Disconnected: ", conn.RemoteAddr())
	conn.Close()
	for index, c := range s.connections {
		if c.RemoteAddr() == conn.RemoteAddr() {
			s.connections = append(s.connections[:index], s.connections[index+1:]...)
			if name != "" {
				s.broadcast(conn, name+" has disconnected")
			}
		}
	}
}

func (s *Server) say(from net.Conn, name, msg string) {
	from.Write([]byte("You: " + msg))
	s.broadcast(from, name+": "+msg)
}

func (s *Server) broadcast(from net.Conn, msg string) {
	fmt.Println(msg)
	for _, c := range s.connections {
		if c.RemoteAddr() != from.RemoteAddr() {
			c.Write([]byte(msg))
		}
	}
}