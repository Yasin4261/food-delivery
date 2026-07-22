package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// wsFrame is the tagged envelope for every frame the server pushes over a chat
// WebSocket. type == "message" carries a new message; type == "read" carries a
// read receipt so the sender's client can show "seen" live (#106). Inbound
// frames (client -> server) are still the bare {"body":...} shape.
type wsFrame struct {
	Type    string          `json:"type"`
	Message *domain.Message `json:"message,omitempty"`
	Read    *readReceipt    `json:"read,omitempty"`
}

// readReceipt announces that ReaderID has read the other party's messages in a
// conversation up to ReadAt.
type readReceipt struct {
	ConversationID int       `json:"conversation_id"`
	ReaderID       int       `json:"reader_id"`
	ReadAt         time.Time `json:"read_at"`
}

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

// isAdmin reports whether the authenticated caller holds the admin role — used
// to admit them as a participant of a support thread.
func isAdmin(claims *service.Claims) bool { return claims.Role == domain.RoleAdmin }

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
	msg, err := h.chat.SendMessage(r.Context(), claims.UserID, isAdmin(claims), id, req.Body)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	h.hub.Broadcast(id, wsFrame{Type: "message", Message: msg})
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
	msgs, total, err := h.chat.Messages(r.Context(), claims.UserID, isAdmin(claims), id, limit, offset)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondPage(w, msgs, limit, offset, total)
}

// MarkRead handles POST /api/v2/chat/conversations/{id}/read (auth) — marks
// the other party's messages as read for the caller.
func (h *ChatHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
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
	if err := h.chat.MarkRead(r.Context(), claims.UserID, isAdmin(claims), id); err != nil {
		respondDomainError(w, err)
		return
	}
	// Tell the other party's live clients that their messages were seen (#106).
	h.hub.Broadcast(id, wsFrame{Type: "read", Read: &readReceipt{
		ConversationID: id,
		ReaderID:       claims.UserID,
		ReadAt:         time.Now().UTC(),
	}})
	w.WriteHeader(http.StatusNoContent)
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
	if _, err := h.chat.Authorize(r.Context(), claims.UserID, isAdmin(claims), id); err != nil {
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
		msg, err := h.chat.SendMessage(r.Context(), claims.UserID, isAdmin(claims), id, in.Body)
		if err != nil {
			continue
		}
		h.hub.Broadcast(id, wsFrame{Type: "message", Message: msg})
	}
}

// ContactSupport handles POST /api/v2/support/conversations (auth) — a user
// opens (or reuses) their own support thread with the platform.
func (h *ChatHandler) ContactSupport(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	conv, err := h.chat.StartSupportConversation(r.Context(), claims.UserID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, conv)
}

type adminSupportRequest struct {
	UserID int `json:"user_id"`
}

// AdminStartSupport handles POST /api/v2/admin/support/conversations (admin) —
// an admin opens (or reuses) a support thread with a target user.
func (h *ChatHandler) AdminStartSupport(w http.ResponseWriter, r *http.Request) {
	var req adminSupportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	conv, err := h.chat.StartSupportConversation(r.Context(), req.UserID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, conv)
}

// AdminListSupport handles GET /api/v2/admin/support/conversations (admin) —
// the support inbox: every support thread, most recently active first.
func (h *ChatHandler) AdminListSupport(w http.ResponseWriter, r *http.Request) {
	convs, err := h.chat.SupportConversations(r.Context())
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondPage(w, convs, len(convs), 0, len(convs))
}
