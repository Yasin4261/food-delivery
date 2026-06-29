package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// ChatHandler exposes the chat REST endpoints and the WebSocket transport.
type ChatHandler struct {
	chat     *service.ChatService
	hub      *Hub
	upgrader websocket.Upgrader
}

// NewChatHandler builds a ChatHandler with its own connection hub.
func NewChatHandler(chat *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chat: chat,
		hub:  NewHub(),
		// Same-origin is not enforced here; tokens authenticate the caller.
		upgrader: websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
}

type startConversationRequest struct {
	ChefID  int  `json:"chef_id"`
	OrderID *int `json:"order_id"`
}

// StartConversation handles POST /api/v2/chat/conversations (auth) — a customer
// opens (or reuses) a thread with a chef.
func (h *ChatHandler) StartConversation(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req startConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	conv, err := h.chat.StartConversation(r.Context(), claims.UserID, req.ChefID, req.OrderID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, conv)
}

// ListConversations handles GET /api/v2/chat/conversations (auth).
func (h *ChatHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	convs, err := h.chat.Conversations(r.Context(), claims.UserID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	// All of the requester's threads are returned (not offset-paginated).
	respondPage(w, convs, len(convs), 0, len(convs))
}

type postMessageRequest struct {
	Body string `json:"body"`
}

// PostMessage handles POST /api/v2/chat/conversations/{id}/messages (auth). It
// persists the message and broadcasts it to any connected WebSocket clients.
func (h *ChatHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}
	var req postMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	msg, err := h.chat.SendMessage(r.Context(), claims.UserID, id, req.Body)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	h.hub.Broadcast(id, msg)
	respondJSON(w, http.StatusCreated, msg)
}

// ListMessages handles GET /api/v2/chat/conversations/{id}/messages (auth).
func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}
	limit, offset := queryInt(r, "limit", 50), queryInt(r, "offset", 0)
	msgs, total, err := h.chat.Messages(r.Context(), claims.UserID, id, limit, offset)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondPage(w, msgs, limit, offset, total)
}

// WebSocket handles GET /api/v2/chat/conversations/{id}/ws (auth). It authorises
// the participant, upgrades the connection, and relays messages live: frames
// received from the socket are persisted and broadcast to the room.
func (h *ChatHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}
	// Authorise before upgrading so failures are plain HTTP responses.
	if _, err := h.chat.Authorize(r.Context(), claims.UserID, id); err != nil {
		respondDomainError(w, err)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return // Upgrade already wrote an error response.
	}
	client := newWSClient(conn)
	h.hub.join(id, client)
	go client.writePump()

	defer func() {
		h.hub.leave(id, client)
		close(client.send)
		_ = conn.Close()
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var in postMessageRequest
		if json.Unmarshal(data, &in) != nil || in.Body == "" {
			continue
		}
		msg, err := h.chat.SendMessage(r.Context(), claims.UserID, id, in.Body)
		if err != nil {
			continue
		}
		h.hub.Broadcast(id, msg)
	}
}
