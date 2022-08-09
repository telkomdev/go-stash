package stash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"testing"
	"time"
)

type Message struct {
	Data string `json:"data"`
}

type Log struct {
	Action  string    `json:"action"`
	Time    time.Time `json:"time"`
	Message Message   `json:"message"`
}

func testConnection(t *testing.T, wg *sync.WaitGroup, host string, port uint64) {
	defer wg.Done()

	_, err := Connect(host, port)
	if err != nil {
		return
	}
	t.Fatal(err)
}

func testWriteData(t *testing.T, wg *sync.WaitGroup, host string, port uint64, opts ...Option) {
	defer wg.Done()

	s, _ := Connect(host, port, opts...)
	defer s.Close()

	logData := Log{
		Action: "get_me",
		Time:   time.Now(),
		Message: Message{
			Data: "get me for me",
		},
	}

	logDataJSON, _ := json.Marshal(logData)

	_, err := s.Write(logDataJSON)
	if err != nil {
		t.Fatal("Cannot write message to host")
	}
}

func testWriteInvalidData(t *testing.T, wg *sync.WaitGroup, host string, port uint64, opts ...Option) {
	defer wg.Done()

	s, _ := Connect(host, port, opts...)

	// early close connection before write data
	s.Close()
	_, err := s.Write(make([]byte, 0))
	if err != nil {
		return
	}
	t.Fatal("Cannot write message to host")
}

func TestStash(t *testing.T) {
	const host string = "localhost"
	const listenPort uint64 = 5000
	const invalidHost string = "localhostnet"
	const invalidListenPort uint64 = 6000

	// Start TCP handler
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	var wg sync.WaitGroup
	wg.Add(4)
	// Test invalid host
	go testConnection(t, &wg, invalidHost, listenPort)
	// Test invalid port
	go testConnection(t, &wg, host, invalidListenPort)
	// Test write
	opts := []Option{
		SetKeepAlive(time.Minute * 1),
		SetReadTimeout(time.Minute * 1),
		SetWriteTimeout(time.Minute * 1),
	}
	go testWriteData(t, &wg, host, listenPort, opts...)
	go testWriteInvalidData(t, &wg, host, listenPort, opts...)
	wg.Wait()

	conn, err := l.Accept()
	defer conn.Close()
	for {
		if err != nil {
			return
		}

		_, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		return
	}
}
