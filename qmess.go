package qmess

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"
)

const (
	RequestConnect = iota
	RequestAccepted
	RequestDisconnect
	RequestPublish
	RequestReceived
	RequestSubscribe
)

type Request struct {
	Type    int
	Sender  string
	Queue   string
	Payload string
}

type Publisher interface {
	Publish(r *Request)
}
type Subscriber interface {
	Subscribe(r *Request)
	Receive(r *Request)
}

type Service struct {
	name   string
	connID int
}

type Queue struct {
	Name        string
	Requests    []Request
	Subscribers []Service
}

type Qmess struct {
	ListenAddr  string
	Connections map[int]net.Conn
	Queues      []Queue
}

func NewServer() *Qmess {
	var m map[int]net.Conn
	m = make(map[int]net.Conn)
	return &Qmess{
		ListenAddr:  "localhost:8428",
		Queues:      []Queue{},
		Connections: m,
	}
}

func (q *Qmess) Run() {
	ln, err := net.Listen("tcp", q.ListenAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("[==] Qmess started [==]")
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
			return
		}
		go q.handleConnection(conn)
	}
}

func (q *Qmess) handleConnection(conn net.Conn) {
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
		connID := rand.Int()
		q.Connections[connID] = conn
		q.handleRequest(&r, connID)
	}
}

func (q *Qmess) handleRequest(r *Request, connID int) {
	switch {
	case r.Type == RequestConnect:
		log.Printf("[+] %s connected (connection ID: %d)[+]\n", r.Sender, connID)
		return
	case r.Type == RequestDisconnect:
		log.Printf("[-] %s disconnected [-]\n", r.Sender)
		return
	case r.Type == RequestPublish:
		q.QueueRoute(*r)
	case r.Type == RequestSubscribe:
		q.AddSub(r, connID)
	}
}

func (q *Qmess) queueExists(qName string) bool {
	for _, q := range q.Queues {
		if q.Name == qName {
			return true
		}
	}
	return false
}

func (q *Qmess) QueueRoute(r Request) {
	if !q.queueExists(r.Queue) {
		q.Queues = append(q.Queues, Queue{Name: r.Queue})
		log.Printf("New queue %s created\n", r.Queue)
	}
	for i, qName := range q.Queues {
		if qName.Name == r.Queue {
			q.Queues[i].Requests = append(q.Queues[i].Requests, r)
			log.Printf("%s published to queue: %s\n", r.Sender, r.Queue)
			return
		}
	}
}

func (q *Qmess) AddSub(r *Request, connID int) {
	if !q.queueExists(r.Queue) {
		q.Queues = append(q.Queues, Queue{Name: r.Queue})
		log.Printf("New queue %s created\n", r.Queue)
	}
	for _, q := range q.Queues {
		if q.Name == r.Queue {
			q.Subscribers = append(q.Subscribers, Service{name: r.Sender, connID: connID})
			log.Printf("%s subscribed to queue %s\n", r.Sender, r.Queue)
			return
		}
	}
}

func (q *Qmess) Push(r Request, s Service) {
	data, err := json.Marshal(&r)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = q.Connections[s.connID].Write(data)
	if err != nil {
		log.Fatal(err)
		return
	}
}

type Client struct {
	name       string
	serverAddr string
	Timeout    time.Duration
	session    net.Conn
}

func NewClient(name string) *Client {
	return &Client{
		name:       name,
		serverAddr: "localhost:8428",
		Timeout:    time.Second * 10,
	}
}

func (c *Client) Run() {
	c.Connect()
}

func (c *Client) Await() {

}

func (c *Client) Send(r *Request) {
	data, err := json.Marshal(&r)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = c.session.Write(data)
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
		Type:   RequestConnect,
		Sender: c.name,
	}
	c.Send(&r)
}

func (c *Client) Close() {
	r := Request{
		Type:   RequestDisconnect,
		Sender: c.name,
	}
	c.Send(&r)
	c.session.Close()
}

func (c *Client) Publish(p string, qName string) {
	r := Request{
		Type:    RequestPublish,
		Sender:  c.name,
		Queue:   qName,
		Payload: p,
	}
	c.Send(&r)
}

func (c *Client) Subscribe(qName string) {
	r := Request{
		Type:   RequestSubscribe,
		Sender: c.name,
		Queue:  qName,
	}
	c.Send(&r)
}

func (c *Client) Received(qName string, mId string) {
	r := Request{
		Type:    RequestReceived,
		Sender:  c.name,
		Queue:   qName,
		Payload: mId,
	}
	c.Send(&r)
}
