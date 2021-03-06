package main

import (
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
	s, err := stash.Connect("localhost", 5000)
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
