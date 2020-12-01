package server

import "log"

// SubPool represents the pool of connections that are present and channels for registering and broadcasting
type SubPool struct {
	Subscribers []*Subscriber
	Register    chan *Subscriber
	Broadcast   chan Message
	StopChan    chan struct{}
}

// NewSubPool creates a new SubPool
func NewSubPool() *SubPool {
	return &SubPool{
		Subscribers: make([]*Subscriber, 0),
		Register:    make(chan *Subscriber),
		Broadcast:   make(chan Message),
		StopChan:    make(chan struct{}),
	}
}

// Start starts the subpool and handles cases for when messages and connections are put onto channels
func (pool *SubPool) Start() {
	for {
		select {
		case sub := <-pool.Register:
			log.Printf("Registering subscriber %v to pool", sub.ID)
			pool.Subscribers = append(pool.Subscribers, sub)
			log.Printf("Successfully registered subscriber %v to pool", sub.ID)
		case message := <-pool.Broadcast:
			log.Printf("Broadcasting message to all subscribers...")
			for _, sub := range pool.Subscribers {
				if err := sub.connection.WriteJSON(message); err != nil {
					log.Printf("Error writing message %v to subscriber %v", message, sub.ID)
				}
			}
		case <-pool.StopChan:
			for _, sub := range pool.Subscribers {
				log.Printf("shutting down connection for subscriber %v...", sub.ID)
				sub.connection.Close()
			}
		}
	}
}

// Stop stops the connections in the sub pool
func (pool *SubPool) Stop() {
	pool.StopChan <- struct{}{}
}
