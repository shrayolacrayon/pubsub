package server

import "net/http"
import "log"
import "fmt"
import "io/ioutil"
import "encoding/json"

import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Message represents a message that is broadcast to all subscribers
type Message struct {
	Body string `json:"body"`
}

// Subscriber represents a client that has a connection to the pool
type Subscriber struct {
	ID         string
	connection *websocket.Conn
	pool       *SubPool
}

// PubSubServer is a server that publishes messages to subscribers
type PubSubServer struct {
	subPool *SubPool
	Address string
	Port    int
	http.Handler
}

// Address is an optional func that will set the host address
func Address(a string) func(p *PubSubServer) error {
	return func(p *PubSubServer) error {
		p.Address = a
		return nil
	}
}

// Port is an optional func that will set the port
func Port(port int) func(p *PubSubServer) error {
	return func(p *PubSubServer) error {
		p.Port = port
		return nil
	}
}

// NewPubSubSerer returns a new server with default to ":8080" with a handler.
func NewPubSubServer(opts ...func(*PubSubServer) error) (*PubSubServer, error) {
	p := &PubSubServer{
		subPool: NewSubPool(),
		Port:    8080,
	}

	for _, f := range opts {
		if err := f(p); err != nil {
			return nil, err
		}
	}

	go p.subPool.Start()

	router := http.NewServeMux()
	router.HandleFunc("/broadcast", p.broadcast)
	router.HandleFunc("/register", p.register)
	p.Handler = router

	return p, nil
}

func (p *PubSubServer) register(w http.ResponseWriter, r *http.Request) {

	// get the connection,  establish a web socket with the client that can be used to broadcast messages
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading connection %v", err)
		http.Error(w, "failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	sub := &Subscriber{ID: conn.RemoteAddr().String(), connection: conn, pool: p.subPool}
	p.subPool.Register <- sub
}

func (p *PubSubServer) broadcast(w http.ResponseWriter, req *http.Request) {
	message := &Message{}
	// get the message contents from the http request
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("error reading request body %s", err), http.StatusBadRequest)
	}
	err = json.Unmarshal(body, message)
	if err != nil {
		http.Error(w, fmt.Sprintf("error unmarshaling request body %s", err), http.StatusBadRequest)
	}
	if message == nil {
		http.Error(w, "failed to include message in the request", http.StatusBadRequest)
	}
	p.subPool.Broadcast <- *message

}

// Stop will stop the sub pool
func (p *PubSubServer) Stop() {
	p.subPool.Stop()
}
