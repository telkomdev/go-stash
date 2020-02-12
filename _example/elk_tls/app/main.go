package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
		os.Exit(1)
	}

	// InsecureSkipVerify: true
	// if CA you are using is a self signed SSL
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

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

	http.HandleFunc("/", Hello(s))
	http.HandleFunc("/log", HelloWithLog(logger))
	http.ListenAndServe(":8080", nil)
}

func Hello(s *stash.Stash) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		fmt.Fprintf(w, "message write count %d!", n)
	}
}

func HelloWithLog(l *log.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logData := Log{
			Action: "get_me",
			Time:   time.Now(),
			Message: Message{
				Data: "get me for me",
			},
		}

		logDataJSON, _ := json.Marshal(logData)

		l.Print(string(logDataJSON))

		fmt.Fprintf(w, "message to log %d!", 0)
	}
}
