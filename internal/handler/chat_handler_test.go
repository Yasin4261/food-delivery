package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeChatRepo is an in-memory domain.ChatRepository for HTTP/WS tests.
type fakeChatRepo struct {
	mu       sync.Mutex
	convs    map[int]*domain.Conversation
	msgs     []*domain.Message
	nextConv int
	nextMsg  int
}

func newFakeChatRepo() *fakeChatRepo {
	return &fakeChatRepo{convs: map[int]*domain.Conversation{}, nextConv: 1, nextMsg: 1}
}

func (f *fakeChatRepo) FindConversation(_ context.Context, userID, chefID int) (*domain.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.convs {
		if c.UserID == userID && c.ChefID == chefID {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrConversationNotFound
}
func (f *fakeChatRepo) FindConversationByID(_ context.Context, id int) (*domain.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if c, ok := f.convs[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, domain.ErrConversationNotFound
}
func (f *fakeChatRepo) CreateConversation(_ context.Context, c *domain.Conversation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c.ID = f.nextConv
	f.nextConv++
	c.CreatedAt = time.Now()
	cp := *c
	f.convs[c.ID] = &cp
	return nil
}
func (f *fakeChatRepo) ListConversationsByUser(_ context.Context, userID int) ([]*domain.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*domain.Conversation, 0)
	for _, c := range f.convs {
		if c.UserID == userID {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakeChatRepo) ListConversationsByChef(_ context.Context, chefID int) ([]*domain.Conversation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*domain.Conversation, 0)
	for _, c := range f.convs {
		if c.ChefID == chefID {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakeChatRepo) CreateMessage(_ context.Context, m *domain.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m.ID = f.nextMsg
	f.nextMsg++
	m.CreatedAt = time.Now()
	cp := *m
	f.msgs = append(f.msgs, &cp)
	if c, ok := f.convs[m.ConversationID]; ok {
		c.LastMessageAt = &m.CreatedAt
	}
	return nil
}
func (f *fakeChatRepo) ListMessages(_ context.Context, conversationID, limit, offset int) ([]*domain.Message, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*domain.Message, 0)
	for _, m := range f.msgs {
		if m.ConversationID == conversationID {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, len(out), nil
}

// startConversation has the customer open a thread with chef 1 and returns the
// conversation id.
func startConversation(t *testing.T, srv http.Handler, customerToken string) int {
	t.Helper()
	rec := do(t, srv, http.MethodPost, "/api/v2/chat/conversations", customerToken, `{"chef_id":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("start conversation = %d (%s)", rec.Code, rec.Body)
	}
	var conv domain.Conversation
	_ = json.Unmarshal(rec.Body.Bytes(), &conv)
	return conv.ID
}

func TestChat_RESTFlowAndAuthorization(t *testing.T) {
	srv := newTestServer()
	chefToken := registerAndToken(t, srv, "chefa", "chefa@example.com")
	createChefProfile(t, srv, chefToken)
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")

	// Unknown chef -> 404.
	if rec := do(t, srv, http.MethodPost, "/api/v2/chat/conversations", customer, `{"chef_id":999}`); rec.Code != http.StatusNotFound {
		t.Errorf("start with unknown chef = %d, want 404", rec.Code)
	}

	convID := startConversation(t, srv, customer)

	// Post a message and load history (persist + load).
	if rec := do(t, srv, http.MethodPost, "/api/v2/chat/conversations/"+itoa(convID)+"/messages", customer, `{"body":"hello chef"}`); rec.Code != http.StatusCreated {
		t.Fatalf("post message = %d (%s)", rec.Code, rec.Body)
	}
	rec := do(t, srv, http.MethodGet, "/api/v2/chat/conversations/"+itoa(convID)+"/messages", customer, "")
	msgs := decodePage[domain.Message](t, rec.Body.Bytes())
	if rec.Code != http.StatusOK || len(msgs.Data) != 1 || msgs.Data[0].Body != "hello chef" || msgs.Total != 1 {
		t.Errorf("history = %d/%+v, want 200 with one message", rec.Code, msgs)
	}

	// The chef (the other participant) can also read it.
	if rec := do(t, srv, http.MethodGet, "/api/v2/chat/conversations/"+itoa(convID)+"/messages", chefToken, ""); rec.Code != http.StatusOK {
		t.Errorf("chef read = %d, want 200", rec.Code)
	}

	// A stranger cannot read or post.
	other := registerCustomerToken(t, srv, "other", "other@example.com")
	if rec := do(t, srv, http.MethodGet, "/api/v2/chat/conversations/"+itoa(convID)+"/messages", other, ""); rec.Code != http.StatusForbidden {
		t.Errorf("stranger read = %d, want 403", rec.Code)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/chat/conversations/"+itoa(convID)+"/messages", other, `{"body":"intrude"}`); rec.Code != http.StatusForbidden {
		t.Errorf("stranger post = %d, want 403", rec.Code)
	}
	// Empty body -> 400.
	if rec := do(t, srv, http.MethodPost, "/api/v2/chat/conversations/"+itoa(convID)+"/messages", customer, `{"body":"  "}`); rec.Code != http.StatusBadRequest {
		t.Errorf("empty body = %d, want 400", rec.Code)
	}
}

// wsURL converts an httptest server URL + path into a ws:// dial URL.
func wsURL(serverURL, path string) string {
	return "ws" + strings.TrimPrefix(serverURL, "http") + path
}

func dialWS(t *testing.T, serverURL, path, token string) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	header := http.Header{"Authorization": {"Bearer " + token}}
	return websocket.DefaultDialer.Dial(wsURL(serverURL, path), header)
}

// TestChat_WebSocketQueryTokenAuth covers the browser path: the WebSocket API
// cannot set an Authorization header, so the token travels as ?access_token=.
func TestChat_WebSocketQueryTokenAuth(t *testing.T) {
	srv := newTestServer()
	chefToken := registerAndToken(t, srv, "chefa", "chefa@example.com")
	createChefProfile(t, srv, chefToken)
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	convID := startConversation(t, srv, customer)

	server := httptest.NewServer(srv)
	defer server.Close()
	base := "/api/v2/chat/conversations/" + itoa(convID) + "/ws"

	// No header, token in the query -> handshake succeeds.
	conn, _, err := websocket.DefaultDialer.Dial(wsURL(server.URL, base+"?access_token="+customer), nil)
	if err != nil {
		t.Fatalf("query-token dial failed: %v", err)
	}
	conn.Close()

	// Garbage query token -> 401 at the handshake.
	if _, resp, err := websocket.DefaultDialer.Dial(wsURL(server.URL, base+"?access_token=garbage"), nil); err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("garbage query token should fail with 401, got err=%v resp=%v", err, resp)
	}
}

func TestChat_WebSocketLiveDelivery(t *testing.T) {
	srv := newTestServer()
	chefToken := registerAndToken(t, srv, "chefa", "chefa@example.com")
	createChefProfile(t, srv, chefToken)
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	convID := startConversation(t, srv, customer)

	server := httptest.NewServer(srv)
	defer server.Close()
	path := "/api/v2/chat/conversations/" + itoa(convID) + "/ws"

	// Both participants connect.
	custConn, _, err := dialWS(t, server.URL, path, customer)
	if err != nil {
		t.Fatalf("customer dial: %v", err)
	}
	defer custConn.Close()
	chefConn, _, err := dialWS(t, server.URL, path, chefToken)
	if err != nil {
		t.Fatalf("chef dial: %v", err)
	}
	defer chefConn.Close()

	// Let both server-side registrations settle, then the customer sends.
	time.Sleep(150 * time.Millisecond)
	if err := custConn.WriteMessage(websocket.TextMessage, []byte(`{"body":"live hello"}`)); err != nil {
		t.Fatalf("write: %v", err)
	}

	// The chef receives it live.
	_ = chefConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := chefConn.ReadMessage()
	if err != nil {
		t.Fatalf("chef read: %v", err)
	}
	var frame struct {
		Type    string          `json:"type"`
		Message *domain.Message `json:"message"`
	}
	if err := json.Unmarshal(data, &frame); err != nil {
		t.Fatalf("decode ws frame: %v", err)
	}
	if frame.Type != "message" || frame.Message == nil {
		t.Fatalf("expected a message frame, got %s", data)
	}
	if frame.Message.Body != "live hello" {
		t.Errorf("ws message body = %q, want %q", frame.Message.Body, "live hello")
	}

	// A non-participant is rejected at the handshake (HTTP 403).
	stranger := registerCustomerToken(t, srv, "other", "other@example.com")
	if _, resp, err := dialWS(t, server.URL, path, stranger); err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Errorf("stranger dial should be 403, got err=%v resp=%v", err, resp)
	}
}

func (f *fakeChatRepo) MarkRead(_ context.Context, conversationID, readerUserID int) error {
	now := time.Now()
	for _, m := range f.msgs {
		if m.ConversationID == conversationID && m.SenderID != readerUserID && m.ReadAt == nil {
			m.ReadAt = &now
		}
	}
	return nil
}

func TestChat_MarkRead(t *testing.T) {
	srv := newTestServer()
	chefToken, _ := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	convID := startConversation(t, srv, customer)
	path := "/api/v2/chat/conversations/" + itoa(convID)

	// Customer sends a message; the chef reads the thread.
	if rec := do(t, srv, http.MethodPost, path+"/messages", customer, `{"body":"hi chef"}`); rec.Code != http.StatusCreated {
		t.Fatalf("post message = %d (%s)", rec.Code, rec.Body)
	}

	// A non-participant cannot mark it read.
	stranger := registerCustomerToken(t, srv, "stranger", "stranger@example.com")
	if rec := do(t, srv, http.MethodPost, path+"/read", stranger, ""); rec.Code != http.StatusForbidden {
		t.Errorf("stranger mark read = %d, want 403", rec.Code)
	}
	// Anonymous -> 401.
	if rec := do(t, srv, http.MethodPost, path+"/read", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("anon mark read = %d, want 401", rec.Code)
	}

	// The chef marks it read -> 204; the customer's message now has read_at.
	if rec := do(t, srv, http.MethodPost, path+"/read", chefToken, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("chef mark read = %d (%s)", rec.Code, rec.Body)
	}
	rec := do(t, srv, http.MethodGet, path+"/messages", chefToken, "")
	msgs := decodePage[domain.Message](t, rec.Body.Bytes())
	if len(msgs.Data) != 1 || msgs.Data[0].ReadAt == nil {
		t.Errorf("message read_at not set after mark read: %+v", msgs.Data)
	}
}

// TestChat_ReadReceiptBroadcast: when one party marks a thread read, the other
// party's live WebSocket receives a "read" frame so it can show "seen" (#106).
func TestChat_ReadReceiptBroadcast(t *testing.T) {
	srv := newTestServer()
	chefToken, _ := seedChefWithItem(t, srv, "chefb", "chefb@example.com")
	customer := registerCustomerToken(t, srv, "custb", "custb@example.com")
	convID := startConversation(t, srv, customer)
	path := "/api/v2/chat/conversations/" + itoa(convID)

	// The customer sends a message, then connects a live socket.
	if rec := do(t, srv, http.MethodPost, path+"/messages", customer, `{"body":"seen me?"}`); rec.Code != http.StatusCreated {
		t.Fatalf("post message = %d (%s)", rec.Code, rec.Body)
	}
	server := httptest.NewServer(srv)
	defer server.Close()
	custConn, _, err := dialWS(t, server.URL, path+"/ws", customer)
	if err != nil {
		t.Fatalf("customer dial: %v", err)
	}
	defer custConn.Close()
	time.Sleep(100 * time.Millisecond) // let the join settle

	// The chef reads the thread (shares the in-process hub via srv).
	if rec := do(t, srv, http.MethodPost, path+"/read", chefToken, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("chef mark read = %d (%s)", rec.Code, rec.Body)
	}

	// The customer's socket receives a read receipt.
	_ = custConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := custConn.ReadMessage()
	if err != nil {
		t.Fatalf("customer read: %v", err)
	}
	var frame struct {
		Type string `json:"type"`
		Read *struct {
			ConversationID int `json:"conversation_id"`
			ReaderID       int `json:"reader_id"`
		} `json:"read"`
	}
	if err := json.Unmarshal(data, &frame); err != nil {
		t.Fatalf("decode ws frame: %v", err)
	}
	if frame.Type != "read" || frame.Read == nil || frame.Read.ConversationID != convID || frame.Read.ReaderID == 0 {
		t.Errorf("expected a read frame for conv %d, got %s", convID, data)
	}
}
