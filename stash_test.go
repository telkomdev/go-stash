package stash

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	if err == nil {
		t.Fatal(err)
	}
	return
}

func testWriteData(t *testing.T, wg *sync.WaitGroup, host string, port uint64, opts ...Option) {
	defer wg.Done()

	s, _ := Connect(host, port, opts...)
	defer s.Close()

	data := Log{
		Action: "get_me",
		Time:   time.Now(),
		Message: Message{
			Data: "get me for me",
		},
	}

	dataJSON, _ := json.Marshal(data)

	_, err := s.Write(dataJSON)
	if err != nil {
		t.Fatal("Cannot write message to host")
	}
	return
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

func handleConnection(conn net.Conn) {
	defer conn.Close()
}

func TestStash(t *testing.T) {
	const host string = "localhost"
	const listenPort uint64 = 5000
	const secureListenPort = 5433
	const invalidHost string = "localhostnet"
	const invalidListenPort uint64 = 6000

	// Start TCP & TLS handler
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		t.Fatal(err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", secureListenPort), tlsConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

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
	tlsOpts := []Option{
		SetTLS(true),
		SetSkipVerify(false),
		SetTLSConfig(tlsConfig),
	}
	go testWriteData(t, &wg, host, secureListenPort, tlsOpts...)
	wg.Wait()

	// Handle TCP connection
	for {
		conn, err := l.Accept()
		if err != nil {
			t.Fatal(err)
			return
		}

		go handleConnection(conn)

		return
	}
}
