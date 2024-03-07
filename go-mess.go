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
	RequestPublish
	RequestSubscribe
)

type Request struct {
	Type    int
	Sender  string
	Payload string
}

type Publisher interface {
	Publish(b []byte)
}
type Subscriber interface {
	Receive(b []byte)
}

type Server struct {
	listenAddr  string
	publishers  []Publisher
	subscribers []Subscriber
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr:  listenAddr,
		publishers:  []Publisher{},
		subscribers: []Subscriber{},
	}
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
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
		handleRequest(&r)
	}
}

func handleRequest(r *Request) {
	switch {
	case r.Type == RequestConnect:
		fmt.Printf("USER %s CONNECTED TO SERVER\n", r.Sender)
		return
	case r.Type == RequestPublish:
		fmt.Printf("%s: %s\n", r.Sender, r.Payload)
		return
	case r.Type == RequestSubscribe:
		fmt.Printf("%s subscribed to %s\n", r.Sender, r.Payload)
		return
	}
}

func main() {
	s := NewServer("localhost:8080")
	s.Run()
}
