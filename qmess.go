package qmess

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

type Qmess struct {
	listenAddr  string
	publishers  []Publisher
	subscribers []Subscriber
	queues      []Queue
}

func NewServer() *Qmess {
	return &Qmess{
		listenAddr:  "localhost:8428",
		publishers:  []Publisher{},
		subscribers: []Subscriber{},
		queues:      []Queue{},
	}
}

func (s *Qmess) AddSub(r *Request) error {
	for _, q := range s.queues {
		if q.name == r.Payload {
			q.subscribers = append(q.subscribers, r.Sender)
			log.Printf("%s subscribed to %s\n", r.Sender, r.Payload)
			return nil
		}
	}
	return fmt.Errorf("Queue %s not found", r.Payload)
}

func (s *Qmess) Run() {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("Qmess started...")
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

func (s *Qmess) Publish(r *Request) {
}

func (s *Qmess) handleConnection(conn net.Conn) {
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

func (s *Qmess) handleRequest(r *Request) {
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

type Client struct {
	name       string
	authKey    string
	serverAddr string
	session    net.Conn
}

func NewClient(name string, serverAddr string) *Client {
	return &Client{
		name:       name,
		serverAddr: serverAddr,
	}
}

func (c *Client) Send(r *Request) {
	data, err := json.Marshal(&r)
	if err != nil {
		log.Fatal(err)
		return
	}
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		log.Fatal(err)
		return
	}

}

func (c *Client) Connect() {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	c.session = conn
	r := Request{
		Type:    RequestConnect,
		Sender:  c.name,
		Payload: c.authKey,
	}
	c.Send(&r)
}

func (c *Client) Close() {
	r := Request{
		Type:    RequestDisconnect,
		Sender:  c.name,
		Payload: c.authKey,
	}
	c.Send(&r)
	c.session.Close()
}

func (c *Client) Subscribe(q string) {
	r := Request{
		Type:    RequestSubscribe,
		Sender:  c.name,
		Payload: q,
	}
	c.Send(&r)
}

func (c *Client) Receive(r *Request) {

}
