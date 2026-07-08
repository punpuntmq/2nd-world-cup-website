package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	StateEvent = "state"
	heartbeat  = 25 * time.Second
)

type Client struct {
	send chan []byte
}

type Event struct {
	Name string
	Data any
}

type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	clients    map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event, 16),
		clients:    make(map[*Client]struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			return
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				close(client.send)
				delete(h.clients, client)
			}
		case event := <-h.broadcast:
			payload, err := encodeEvent(event)
			if err != nil {
				log.Printf("sse encode %s failed: %v", event.Name, err)
				continue
			}
			for client := range h.clients {
				select {
				case client.send <- payload:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(eventName string, data any) {
	select {
	case h.broadcast <- Event{Name: eventName, Data: data}:
	default:
		log.Printf("sse broadcast queue full; dropping %s event", eventName)
	}
}

func (h *Hub) BroadcastState(state any) {
	h.Broadcast(StateEvent, state)
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, initialState any) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	header := w.Header()
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	header.Set("Cache-Control", "no-cache, no-transform")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")

	client := &Client{send: make(chan []byte, 8)}
	select {
	case h.register <- client:
	case <-r.Context().Done():
		return
	}
	defer func() {
		select {
		case h.unregister <- client:
		case <-r.Context().Done():
		}
	}()

	if initialState != nil {
		payload, err := encodeEvent(Event{Name: StateEvent, Data: initialState})
		if err != nil {
			http.Error(w, "could not encode initial state", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(payload); err != nil {
			return
		}
		flusher.Flush()
	}

	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case payload, ok := <-client.send:
			if !ok {
				return
			}
			if _, err := w.Write(payload); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := w.Write([]byte(": heartbeat\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func encodeEvent(event Event) ([]byte, error) {
	name := event.Name
	if name == "" {
		name = "message"
	}
	data, err := json.Marshal(event.Data)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", name, data)), nil
}
