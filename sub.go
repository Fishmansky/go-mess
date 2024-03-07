package main

import (
	"encoding/json"
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

type Client struct {
	name       string
	serverAddr string
}

func NewClient(name string, serverAddr string) *Client {
	return &Client{
		name:       name,
		serverAddr: serverAddr,
	}
}

func (c *Client) Send(r *Request) {
	data, err := json.Marshal(r)
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
	r := Request{
		Type:    RequestConnect,
		Sender:  c.name,
		Payload: "password:123123;role=publisher",
	}
	c.Send(&r)
}

func (c *Client) Publish() {
	r := Request{
		Type:    RequestPublish,
		Sender:  c.name,
		Payload: "Stormy weather is to be expected",
	}
	c.Send(&r)
}

func (c *Client) Subsribe() {
	r := Request{
		Type:    RequestSubscribe,
		Sender:  c.name,
		Payload: "weather-report",
	}
	c.Send(&r)
}

func main() {
	c := NewClient("Sub", "localhost:8080")
	c.Subsribe()
}
