package main

import (
	"encoding/json"
	"log"
	"net"
	"time"
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

func main() {
	c := NewClient("Con", "localhost:8080")
	c.Connect()
	time.Sleep(time.Millisecond * 2000)
	c.Subscribe("weather-report")
}
