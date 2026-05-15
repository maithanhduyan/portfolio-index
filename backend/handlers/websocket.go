package handlers

import (
	"context"
	"log"
	"sync"

	fiberws "github.com/gofiber/websocket/v2"

	"portfolio-index/services"
)

// client represents a single WebSocket connection.
type client struct {
	conn *fiberws.Conn
	send chan []byte
	mu   sync.Mutex
}

func (cl *client) write(msg []byte) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	select {
	case cl.send <- msg:
	default:
		// Drop message if buffer full — prevents slow client from blocking
	}
}

// Hub manages all WebSocket clients and broadcasts via Redis pub/sub.
type Hub struct {
	clients    map[*client]struct{}
	register   chan *client
	unregister chan *client
	mu         sync.RWMutex
	cache      *services.RedisCache
}

func NewHub(cache *services.RedisCache) *Hub {
	return &Hub{
		clients:    make(map[*client]struct{}),
		register:   make(chan *client, 256),
		unregister: make(chan *client, 256),
		cache:      cache,
	}
}

// Run listens for Redis pub/sub messages and dispatches to clients.
// Must be run in a dedicated goroutine.
func (h *Hub) Run() {
	ctx := context.Background()
	pubsub := h.cache.Subscribe(ctx, "price_updates")
	defer pubsub.Close()

	redisCh := pubsub.Channel()

	for {
		select {
		case cl := <-h.register:
			h.mu.Lock()
			h.clients[cl] = struct{}{}
			h.mu.Unlock()
			log.Printf("[hub] client connected  | total=%d", len(h.clients))

		case cl := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[cl]; ok {
				delete(h.clients, cl)
				close(cl.send)
			}
			h.mu.Unlock()
			log.Printf("[hub] client disconnected | total=%d", len(h.clients))

		case msg, ok := <-redisCh:
			if !ok {
				return
			}
			payload := []byte(msg.Payload)
			h.mu.RLock()
			for cl := range h.clients {
				cl.write(payload)
			}
			h.mu.RUnlock()
		}
	}
}

// HandleConnection is the Fiber WebSocket handler for each new connection.
func (h *Hub) HandleConnection(c *fiberws.Conn) {
	cl := &client{
		conn: c,
		send: make(chan []byte, 512),
	}

	h.register <- cl

	defer func() {
		h.unregister <- cl
		c.Close()
	}()

	// Write pump — sends buffered messages to WebSocket client.
	go func() {
		for msg := range cl.send {
			if err := c.WriteMessage(fiberws.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	// Read pump — keeps connection alive and handles ping/pong.
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
