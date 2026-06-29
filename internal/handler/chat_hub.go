package handler

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub fans messages out to the WebSocket clients connected to each conversation
// (keyed by conversation id). It is safe for concurrent use.
type Hub struct {
	mu    sync.RWMutex
	rooms map[int]map[*wsClient]struct{}
}

// NewHub builds an empty Hub.
func NewHub() *Hub {
	return &Hub{rooms: map[int]map[*wsClient]struct{}{}}
}

// join registers a client in a conversation room.
func (h *Hub) join(conversationID int, c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[conversationID]
	if room == nil {
		room = map[*wsClient]struct{}{}
		h.rooms[conversationID] = room
	}
	room[c] = struct{}{}
}

// leave removes a client from a conversation room.
func (h *Hub) leave(conversationID int, c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room := h.rooms[conversationID]; room != nil {
		delete(room, c)
		if len(room) == 0 {
			delete(h.rooms, conversationID)
		}
	}
}

// Broadcast sends payload (marshalled to JSON) to every client in the
// conversation room. Slow clients whose buffer is full are skipped.
func (h *Hub) Broadcast(conversationID int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[conversationID] {
		select {
		case c.send <- data:
		default: // drop for an unresponsive client
		}
	}
}

// wsClient is one WebSocket connection with a buffered outbound queue.
type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

func newWSClient(conn *websocket.Conn) *wsClient {
	return &wsClient{conn: conn, send: make(chan []byte, 16)}
}

// writePump drains the send queue to the socket until it is closed.
func (c *wsClient) writePump() {
	for data := range c.send {
		_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}
