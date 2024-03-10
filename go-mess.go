package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	RequestConnect = iota
	RequestDisconnect
	RequestPublish
	RequestSubscribe
)

type Request struct {
	Type    int
	Sender  string
	Payload string
}

type Publisher interface {
	Publish(r *Request)
}
type Subscriber interface {
	Subscribe(r *Request)
	Receive(r *Request)
}

type Queue struct {
	name        string
	subscribers []string
}

type Server struct {
	listenAddr  string
	publishers  []Publisher
	subscribers []Subscriber
	queues      []Queue
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr:  listenAddr,
		publishers:  []Publisher{},
		subscribers: []Subscriber{},
		queues:      []Queue{},
	}
}

func (s *Server) AddSub(r *Request) error {
	for _, q := range s.queues {
		if q.name == r.Payload {
			q.subscribers = append(q.subscribers, r.Sender)
			log.Printf("%s subscribed to %s\n", r.Sender, r.Payload)
			return nil
		}
	}
	return fmt.Errorf("Queue %s not found", r.Payload)
}

func (s *Server) Run() {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("Server started...")
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
			return
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) Publish(r *Request) {
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading from connection:", err)
			}
			break
		}
		var r Request
		json.Unmarshal(buffer[:n], &r)
		s.handleRequest(&r)
	}
}

func (s *Server) handleRequest(r *Request) {
	switch {
	case r.Type == RequestConnect:
		fmt.Printf("USER %s CONNECTED TO SERVER\n", r.Sender)
		return
	case r.Type == RequestDisconnect:
		fmt.Printf("USER %s DISCONNECTED\n", r.Sender)
		return
	case r.Type == RequestPublish:
		fmt.Printf("%s published: %s\n", r.Sender, r.Payload)
		return
	case r.Type == RequestSubscribe:
		err := s.AddSub(r)
		if err != nil {
			log.Println(err)
		}
		return
	}
}

func main() {
	s := NewServer("localhost:8080")
	s.Run()
}
