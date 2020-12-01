package server

import "testing"
import "fmt"
import "net/http"
import "github.com/gorilla/websocket"
import "encoding/json"
import "bytes"
import "io/ioutil"
import "time"

func startServer() {
	fmt.Println("starting server...")
	s := NewPubSubServer()
	defer s.Stop()

	http.ListenAndServe(fmt.Sprintf("%s:%d", s.Address, s.Port), s.Handler)
}

func createSubscriber() (*websocket.Conn, error) {
	dialer := websocket.Dialer{}
	header := http.Header{}

	conn, resp, err := dialer.Dial(fmt.Sprintf("ws://%s:%d/register", "localhost", 5000), header)
	if err != nil {
		return nil, fmt.Errorf("error creating dialer: %s", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, fmt.Errorf("expected status to be Switch Protocols but was %v", resp.StatusCode)
	}

	return conn, nil
}

func broadcastMessage(message string) error {
	client := http.Client{}
	msg := Message{message}
	msgBytes, err := json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("error marshaling message %v", msg)
	}

	req, err := http.NewRequest("POST", "http://localhost:5000/broadcast", bytes.NewBuffer(msgBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)

	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making http request: %v", err)

	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body %v", err)

	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %q", resp.StatusCode, body)

	}
	return nil

}

func TestPubSub(t *testing.T) {
	go startServer()

	time.Sleep(1 * time.Second)
	conns := []*websocket.Conn{}
	for i := 0; i < 10; i++ {
		conn, err := createSubscriber()
		if err != nil {
			t.Errorf("error creating subscriber: %s", err)
			return
		}
		defer func() {
			err := conn.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()
		conns = append(conns, conn)
	}

	messages := []string{"first message", "second message", "third message"}

	for _, message := range messages {
		err := broadcastMessage(message)
		if err != nil {
			t.Errorf("error broadcasting: %s", err)
			return
		}

		for _, conn := range conns {
			received := &Message{}
			conn.ReadJSON(&received)
			if received.Body != message {
				t.Errorf("expected message body %s but got %s", message, received.Body)
			}

		}
	}

}
