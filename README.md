## Go Stash

Logstash Client Library For Go

[![GoDoc](https://godoc.org/github.com/telkomdev/go-stash?status.svg)](https://godoc.org/github.com/telkomdev/go-stash)

### Getting started

Install

```shell
$ go get github.com/telkomdev/go-stash
```


Usage

<b>Basic</b>
```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/telkomdev/go-stash"
)

type Message struct {
	Data string `json:"data"`
}

type Log struct {
	Action  string    `json:"action"`
	Time    time.Time `json:"time"`
	Message Message   `json:"message"`
}

func main() {
    var host string = "localhost"
	var port uint64 = 5000
	s, err := stash.Connect(host, port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		s.Close()
	}()

	logData := Log{
			Action: "get_me",
			Time:   time.Now(),
			Message: Message{
				Data: "get me for me",
			},
		}

    logDataJSON, _ := json.Marshal(logData)

    n, err := s.Write(logDataJSON)
    if err != nil {
        fmt.Fprintf(w, err.Error())
        return
    }
}

```

<b>You can use `go-stash` With Golang's standar Log</b>

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/telkomdev/go-stash"
)

type Message struct {
	Data string `json:"data"`
}

type Log struct {
	Action  string    `json:"action"`
	Time    time.Time `json:"time"`
	Message Message   `json:"message"`
}

func main() {
	var host string = "localhost"
	var port uint64 = 5000
	s, err := stash.Connect(host, port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		s.Close()
	}()

	logger := log.New(s, "", 0)

	logData := Log{
		Action: "get_me",
		Time:   time.Now(),
		Message: Message{
			Data: "get me for me",
		},
	}

	logDataJSON, _ := json.Marshal(logData)

	logger.Print(string(logDataJSON))

}

```

### TLS
```go
func main() {
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
		os.Exit(1)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}

	var (
		host string = "localhost"
		port uint64 = 5000
	)

	s, err := stash.Connect(host, port, stash.SetTLSConfig(&config), stash.SetTLS(true))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		s.Close()
	}()

	logger := log.New(s, "", 0)
```