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
	cer, err := tls.X509KeyPair([]byte("-----BEGIN CERTIFICATE-----\nMIIFszCCA5ugAwIBAgIUBZNBpLJtsBRZMJ5pOE/5Rnm2454wDQYJKoZIhvcNAQEL\nBQAwaTELMAkGA1UEBhMCSUQxEzARBgNVBAgMCkphd2EgQmFyYXQxEDAOBgNVBAcM\nB0pha2FydGExJTAjBgNVBAoMHFBULiBUZWxla29tdW5pa2FzaSBJbmRvbmVzaWEx\nDDAKBgNVBAsMA0VXWjAeFw0yMjA4MDkwODA1MTVaFw0zMjA4MDYwODA1MTVaMGkx\nCzAJBgNVBAYTAklEMRMwEQYDVQQIDApKYXdhIEJhcmF0MRAwDgYDVQQHDAdKYWth\ncnRhMSUwIwYDVQQKDBxQVC4gVGVsZWtvbXVuaWthc2kgSW5kb25lc2lhMQwwCgYD\nVQQLDANFV1owggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDOdSF/7N11\nCYo/ZeBRfiKCd7rNzAOsU7CM2G+ruCfXxkmHfhIX3dHFG4H8OXipSGC1pF8NcQfB\naSNzM7bxew2NMTXhvOG/TqgNM6Kxgbt6URTWQRbiv3+3w5dG+2hCRuXyFh45ygGo\nZvNs5gYsI6sJ5lw2fNsQipgvoi1yAJVqRkINN6I0B7mTJfZ5hzrBUwu1PGPXs1nz\n8cH3sa1O5P3WKdJFOXAXJbk2iBTPxYvZpcW4BB5g+Lo4y1dt4aDja8wEg1nn/729\niLsodVQ7RbCdh6txGYwREFS+gxUswWzbRXzFyJbsZLmWIer/X/+1S//2VzwJmZ25\nSj4C8dzPWmi8vTBTnbcoBm/er3fBe6spc8CV8bucE6tcT096rCenCGHfJfHEDXP6\n60GPeIEJP5FBZIvCN51fMWQW5QUL689is1BfcRhxgICBELkv4heg9QVvDqA+BKZF\n5IYZEvN3bLNL2TnQqQ0jvoh48hQISckZX8TV2fQsv/qiqhebVd3FOtQLNWa6NivP\nZo4Mz7L9BzCIs1xfk7mFjClIDf3oRUw7Gx6hDolAFi3wE/UCT4XjxYq5+X3e/w3R\nVotJXH/nLIuNiFmYVbW5iJiy0L4QClZ/ZdqSls1QzMxdManRIU5V6KzKb5Ns/23e\nmfGxb2GCt5ebN9nNHpkV99JHWA3YqSuUuwIDAQABo1MwUTAdBgNVHQ4EFgQUyVBx\n1gkokmpntkI3XxZ85w786F4wHwYDVR0jBBgwFoAUyVBx1gkokmpntkI3XxZ85w78\n6F4wDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAgEAY64DP5o+CCQG\nPeZch9KxS9/HwUH4ihyPMFq+w+3tfGN/wqpwE7SpATUasPyn5/OfsiLHdtH8q0jt\nTgJ35bEVvDJvW+a/UhX/g63pR05c5l4PtdcIHrBg3t+/7gCoVbaRmFQgTf6m9zBI\nbyNr+JkBLPSiE7wL7lw2KqwJ5Po7wz0jAsPebUyNnfJUuPMaURaaoLKDbqx/QF2/\n3q41A07HJKf8TiibormVM+ZD8W/tPU20lg/P6vDXcQZXoPIsdHJNtZF4tKW0qaGE\nDYG7BRNqEjHql+GZ63GBXB+elzUSISpMXextTnDFxLABZSkZgSxp7GSXfT/FXQIM\nE0Eb5TgoNyYVNBHvU+RfXQQQBc3PvQbG6rB93Zdm8SZ1mxyXymvOASxY4/UAX7K4\nl3PT0asoryK0NqcJZfD9GCIF0NFsCKHgu9EF9bZldplb2kEm0lMPCMyL7ijZxAn+\nHyo3yD+JU/LW6BNv/3HkcUrUXRlVMlIZOF6qfzmz+VDQTpsiOWkRnLxdlq5WuDRa\n/WSSrOP7QBZIxS6cigpjoFiUujLVqG3tL5rPHsSc0tXds3SJNTcEqCEELgRW0rAm\n08LdSZASvOUYtkZCfX4l74uryrYg5vOZUjRN95fTQ+0HTgB5ZG+UX57XoLxLO/Pt\nLby3mPQlMJprxdQJ0ALuR8rVti/3lm0=\n-----END CERTIFICATE-----\n"), []byte("-----BEGIN PRIVATE KEY-----\n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\nRldjXOjCE69qcxZzyWxPCb5gedmuDA==\n-----END PRIVATE KEY-----\n"))
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
