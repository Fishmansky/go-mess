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

Then use Go command to run it:
```bash
$ go run example
```
