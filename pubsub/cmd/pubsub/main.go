package main

import server "pubsub/pkg/server"
import "log"
import "net/http"
import "fmt"
import "os"

func main() {
	log.Printf("Starting server...")
	server, err := server.NewPubSubServer(server.Address("localhost"), server.Port(5000))
	if err != nil {
		os.Exit(1)
	}
	defer server.Stop()

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", server.Address, server.Port), server.Handler))
}
