package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
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

	// InsecureSkipVerify: true if CA you are using self signed SSL
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

	var logger = log.New()

	// Log as JSON instead of the default ASCII formatter.
	logger.Formatter = &log.JSONFormatter{}

	logger.Out = s

	logger.Level = log.InfoLevel

	http.HandleFunc("/", HelloWithLogrus(logger))
	http.ListenAndServe(":8080", nil)
}

func HelloWithLogrus(l *log.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l.WithFields(log.Fields{
			"action": "get_me",
			"time":   time.Now(),
			"message": map[string]interface{}{
				"data": "get me for me",
			},
		}).Info()

		fmt.Fprintf(w, "message to log %d!", 0)
	}
}
