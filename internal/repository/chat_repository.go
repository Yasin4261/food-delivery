package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// ChatRepository is the PostgreSQL adapter for domain.ChatRepository.
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository builds a ChatRepository.
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

const conversationColumns = `id, user_id, chef_id, order_id, last_message_at, created_at`

func scanConversation(s interface{ Scan(...any) error }) (*domain.Conversation, error) {
	c := &domain.Conversation{}
	err := s.Scan(&c.ID, &c.UserID, &c.ChefID, &c.OrderID, &c.LastMessageAt, &c.CreatedAt)
	return c, err
}

// FindConversation returns the thread for a (userID, chefID) pair.
func (r *ChatRepository) FindConversation(ctx context.Context, userID, chefID int) (*domain.Conversation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+conversationColumns+` FROM chat_conversations WHERE user_id = $1 AND chef_id = $2`, userID, chefID)
	c, err := scanConversation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrConversationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find conversation: %w", err)
	}
	return c, nil
}

// FindConversationByID returns the conversation with the given id.
func (r *ChatRepository) FindConversationByID(ctx context.Context, id int) (*domain.Conversation, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+conversationColumns+` FROM chat_conversations WHERE id = $1`, id)
	c, err := scanConversation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrConversationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find conversation by id: %w", err)
	}
	return c, nil
}

// CreateConversation inserts a thread and back-fills id and created_at.
func (r *ChatRepository) CreateConversation(ctx context.Context, c *domain.Conversation) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO chat_conversations (user_id, chef_id, order_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`, c.UserID, c.ChefID, c.OrderID).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return fmt.Errorf("create conversation: %w", err)
	}
	return nil
}

// In a conversation only the customer (c.user_id) and the chef send messages,
// so a side's unread count is expressible purely from c.user_id — no need to
// know the requester's user id.
const unreadForCustomer = `(SELECT count(*) FROM chat_messages m
	WHERE m.conversation_id = c.id AND m.sender_id <> c.user_id AND m.read_at IS NULL)`
const unreadForChef = `(SELECT count(*) FROM chat_messages m
	WHERE m.conversation_id = c.id AND m.sender_id = c.user_id AND m.read_at IS NULL)`

// ListConversationsByUser returns a customer's threads, most recently active
// first, each with the customer's unread count.
func (r *ChatRepository) ListConversationsByUser(ctx context.Context, userID int) ([]*domain.Conversation, error) {
	return r.listConversations(ctx, `
		SELECT `+conversationColumns+`, `+unreadForCustomer+` AS unread
		FROM chat_conversations c WHERE c.user_id = $1
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC`, userID)
}

// ListConversationsByChef returns a chef's threads, most recently active
// first, each with the chef's unread count.
func (r *ChatRepository) ListConversationsByChef(ctx context.Context, chefID int) ([]*domain.Conversation, error) {
	return r.listConversations(ctx, `
		SELECT `+conversationColumns+`, `+unreadForChef+` AS unread
		FROM chat_conversations c WHERE c.chef_id = $1
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC`, chefID)
}

func (r *ChatRepository) listConversations(ctx context.Context, query string, arg int) ([]*domain.Conversation, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.Conversation, 0)
	for rows.Next() {
		c := &domain.Conversation{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.ChefID, &c.OrderID, &c.LastMessageAt, &c.CreatedAt, &c.UnreadCount); err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// MarkRead stamps read_at on messages in the conversation not sent by the
// reader that are still unread.
func (r *ChatRepository) MarkRead(ctx context.Context, conversationID, readerUserID int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE chat_messages SET read_at = now()
		WHERE conversation_id = $1 AND sender_id <> $2 AND read_at IS NULL`, conversationID, readerUserID)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

// CreateMessage inserts a message and bumps the conversation's last_message_at
// in one transaction.
func (r *ChatRepository) CreateMessage(ctx context.Context, m *domain.Message) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO chat_messages (conversation_id, sender_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`, m.ConversationID, m.SenderID, m.Body).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE chat_conversations SET last_message_at = $2 WHERE id = $1`, m.ConversationID, m.CreatedAt); err != nil {
		return fmt.Errorf("bump last_message_at: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit message: %w", err)
	}
	return nil
}

// ListMessages returns a page of a conversation's messages (oldest first) and
// the total message count.
func (r *ChatRepository) ListMessages(ctx context.Context, conversationID, limit, offset int) ([]*domain.Message, int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, conversation_id, sender_id, body, read_at, created_at
		FROM chat_messages WHERE conversation_id = $1
		ORDER BY created_at ASC, id ASC
		LIMIT $2 OFFSET $3`, conversationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.Message, 0)
	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Body, &m.ReadAt, &m.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan message: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM chat_messages WHERE conversation_id = $1`, conversationID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count messages: %w", err)
	}
	return out, total, nil
}
