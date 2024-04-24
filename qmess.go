package qmess

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
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
	Subscribers []Service

	muRequests sync.Mutex
	Requests   []Request
}

type Qmess struct {
	ListenAddr string
	Queues     []Queue

	muConnections sync.Mutex
	Connections   map[int]*net.Conn
}

func NewServer() *Qmess {
	var m map[int]*net.Conn
	m = make(map[int]*net.Conn)
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
		connID := rand.Intn(1000)
		go q.handleConnection(conn, connID)
	}
}

func (q *Qmess) handleConnection(conn net.Conn, connID int) {
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
		q.handleRequest(&r, &conn, connID)
	}
}

func (q *Qmess) handleRequest(r *Request, conn *net.Conn, connID int) {
	switch {
	case r.Type == RequestConnect:
		q.NewConnection(*r, conn, connID)
	case r.Type == RequestDisconnect:
		q.Disconnect(*r, connID)
	case r.Type == RequestPublish:
		q.QueueRoute(*r, connID)
	case r.Type == RequestSubscribe:
		q.Subscribe(*r, connID)
	}
}

func (q *Qmess) NewConnection(r Request, conn *net.Conn, connID int) {
	resp := &Request{
		Type:   RequestAccepted,
		Sender: "qmess-server",
	}
	data, err := json.Marshal(&resp)
	if err != nil {
		log.Println("Response marshalling error:", err)
	}
	q.muConnections.Lock()
	c := *conn
	q.Connections[connID] = conn
	_, err = c.Write([]byte(data))
	if err != nil {
		log.Println("Connection request error:", err)
	}
	log.Printf("[+] %s connected (connection ID: %d)[+]\n", r.Sender, connID)
	q.muConnections.Unlock()
}

func (q *Qmess) Disconnect(r Request, connID int) {
	q.muConnections.Lock()
	conn := *q.Connections[connID]
	log.Printf("[-] Connection ID %d closed [-]\n", connID)
	conn.Close()
	delete(q.Connections, connID)
	log.Printf("[-] %s disconnected [-]\n", r.Sender)
	q.muConnections.Unlock()
}

func (q *Qmess) queueExists(qName string) bool {
	for _, q := range q.Queues {
		if q.Name == qName {
			return true
		}
	}
	return false
}

func (q *Qmess) QueueRoute(r Request, connID int) {
	if !q.queueExists(r.Queue) {
		q.Queues = append(q.Queues, Queue{Name: r.Queue})
		log.Printf("New queue %s created\n", r.Queue)
	}
	for i, qu := range q.Queues {
		if qu.Name == r.Queue {
			q.Queues[i].Requests = append(q.Queues[i].Requests, r)
			log.Printf("%s (connID: %d) published to queue: %s\n", r.Sender, connID, r.Queue)
		}
	}
}

func (q *Qmess) Subscribe(r Request, connID int) {
	if !q.queueExists(r.Queue) {
		q.Queues = append(q.Queues, Queue{Name: r.Queue})
		log.Printf("New queue %s created\n", r.Queue)
	}
	for i, qu := range q.Queues {
		if qu.Name == r.Queue {
			log.Printf("%s subscribed to queue %s\n", r.Sender, r.Queue)
			for {
				time.Sleep(time.Millisecond * 200)
				q.PushFirst(i, connID)
			}
		}
	}
}

func (q *Qmess) PushFirst(qID int, connID int) {
	if len(q.Queues[qID].Requests) > 0 {
		q.Queues[qID].muRequests.Lock()
		q.muConnections.Lock()
		data, err := json.Marshal(q.Queues[qID].Requests[0])
		if err != nil {
			log.Fatal(err)
			return
		}
		ses := *q.Connections[connID]
		_, err = ses.Write(data)
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Printf("Request pushed to connection ID %d", connID)
		q.Queues[qID].Requests = q.Queues[qID].Requests[1:]
		q.muConnections.Unlock()
		q.Queues[qID].muRequests.Unlock()
	}
}

type Client struct {
	name       string
	serverAddr string
	Timeout    time.Duration
	Session    *net.Conn
}

func NewClient(name string) *Client {
	return &Client{
		name:       name,
		serverAddr: "localhost:8428",
		Timeout:    time.Second * 10,
	}
}

func (c *Client) Send(r *Request) {
	data, err := json.Marshal(&r)
	if err != nil {
		log.Fatal(err)
		return
	}
	s := *c.Session
	_, err = s.Write(data)
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
	c.Session = &conn
	r := Request{
		Type:   RequestConnect,
		Sender: c.name,
	}
	c.Send(&r)
}

func (c *Client) RequestResponse() {
	conn := *c.Session
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading from conection:", err)
		}
		return
	}
	var resp Request
	json.Unmarshal(buffer[:n], &resp)
	if resp.Type != RequestAccepted {
		log.Fatal("Connection response other than accepted - exiting now.")
	}
}

func (c *Client) Close() {
	r := Request{
		Type:   RequestDisconnect,
		Sender: c.name,
	}
	c.Send(&r)
	s := *c.Session
	s.Close()
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
	for {
		conn := *c.Session
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading from conection:", err)
			}
			break
		}
		var r Request
		json.Unmarshal(buffer[:n], &r)
		fmt.Println(r)
	}
}
