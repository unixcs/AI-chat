package store

import (
	"database/sql"
	"errors"

	"ai-chat-backend/internal/model"
)

// ---------- conversations ----------

const convColumns = `id, userId, title, createdAt, updatedAt`

func scanConversation(row interface{ Scan(...any) error }) (*model.Conversation, error) {
	c := &model.Conversation{}
	err := row.Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) ListConversationsByUser(userID string, limit, offset int) ([]model.Conversation, error) {
	rows, err := s.DB.Query(`SELECT `+convColumns+` FROM conversations WHERE userId = ? ORDER BY updatedAt DESC LIMIT ? OFFSET ?`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Conversation{}
	for rows.Next() {
		c, err := scanConversation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (s *Store) GetConversation(id string) (*model.Conversation, error) {
	c, err := scanConversation(s.DB.QueryRow(`SELECT `+convColumns+` FROM conversations WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *Store) CreateConversation(c *model.Conversation) error {
	_, err := s.DB.Exec(`INSERT INTO conversations (id, userId, title, createdAt, updatedAt) VALUES (?,?,?,?,?)`,
		c.ID, c.UserID, c.Title, c.CreatedAt, c.UpdatedAt)
	return err
}

func (s *Store) TouchConversation(id string) error {
	_, err := s.DB.Exec(`UPDATE conversations SET updatedAt = ? WHERE id = ?`, Now(), id)
	return err
}

// DeleteConversation removes the conversation and its messages atomically.
func (s *Store) DeleteConversation(id string) error {
	return s.inTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`DELETE FROM messages WHERE conversationId = ?`, id); err != nil {
			return err
		}
		res, err := tx.Exec(`DELETE FROM conversations WHERE id = ?`, id)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// ---------- messages ----------

func (s *Store) GetMessages(conversationID string) ([]model.Message, error) {
	rows, err := s.DB.Query(`SELECT id, conversationId, role, content, createdAt FROM messages WHERE conversationId = ? ORDER BY createdAt`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Message{}
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) InsertMessage(m *model.Message) error {
	_, err := s.DB.Exec(`INSERT INTO messages (id, conversationId, role, content, createdAt) VALUES (?,?,?,?,?)`,
		m.ID, m.ConversationID, m.Role, m.Content, m.CreatedAt)
	return err
}

// SearchConversationIDsByContent returns IDs of conversations containing the keyword.
func (s *Store) SearchConversationIDsByContent(keyword string) (map[string]bool, error) {
	rows, err := s.DB.Query(`SELECT DISTINCT conversationId FROM messages WHERE lower(content) LIKE '%' || ? || '%'`, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
