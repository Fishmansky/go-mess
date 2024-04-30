# qmess
TCP based message broker for microservices

# Usage

## Installation

Install `qmess` package:
```bash
$ go get -u github.com/Fishmansky/qmess
```

## Example

Create `example.go` with content like below:

```go
package main

import "github.com/Fishmansky/qmess"

func main() {
	q := qmess.NewServer()
	q.Run() // listen and server on localhost:8428
}
```
Message broker will listen and serve simple message queue

To publish messages to queue create client, prepare data to publish and specify queue to publish to:

```go
func main() {
	c := qmess.NewClient("order-creator")
	c.Connect()
	// app logic

	// send data to specified queue
	c.Publish("order nr 1", "orders")

	// close connection
	c.Close()
}
```

Subcribe to queue defining client, connecting and specifying queue:

```go
func main() {
	c := qmess.NewClient("order-receiver")
	c.Connect()
	c.Subscribe("orders")
}
```

Run broker:
```bash
$ go run example.go
```
