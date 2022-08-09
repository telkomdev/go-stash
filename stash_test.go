package stash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
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

func TestStash(t *testing.T) {
	const host string = "localhost"
	const listenPort uint64 = 5000
	timeNow := time.Now()

	go func() {
		s, err := Connect(host, listenPort)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer s.Close()

		logData := Log{
			Action: "get_me",
			Time:   timeNow,
			Message: Message{
				Data: "get me for me",
			},
		}

		logDataJSON, _ := json.Marshal(logData)

		_, err = s.Write(logDataJSON)
		if err != nil {
			t.Fatal("Cannot write message to host")
		}
	}()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}
		respData := Log{
			Action: "get_me",
			Time:   timeNow,
			Message: Message{
				Data: "get me for me",
			},
		}
		reqData := Log{}
		_ = json.Unmarshal(buf[:], &reqData)

		if reqData.Message != respData.Message {
			t.Fatalf("Unexpected message:\nGot:\t\t%s\nExpected:\t%s\n", respData.Message, reqData.Message)
		}

		return
	}
}
