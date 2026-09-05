package store

import (
	"database/sql"
	"errors"
	"fmt"

	"ai-chat-backend/internal/model"
)

// ---------- redeem codes ----------

func scanCode(row interface{ Scan(...any) error }) (*model.RedeemCode, error) {
	c := &model.RedeemCode{}
	var usedAt, usedBy sql.NullString
	err := row.Scan(&c.ID, &c.Code, &c.DurationMonths, &c.Status, &c.CreatedAt, &usedAt, &usedBy)
	if err != nil {
		return nil, err
	}
	c.UsedAt = usedAt.String
	c.UsedByUserID = usedBy.String
	return c, nil
}

func (s *Store) GetRedeemCodeByCode(code string) (*model.RedeemCode, error) {
	c, err := scanCode(s.DB.QueryRow(`SELECT id, code, durationMonths, status, createdAt, usedAt, usedByUserId FROM redeemCodes WHERE code = ?`, code))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *Store) GetRedeemCodeByID(id string) (*model.RedeemCode, error) {
	c, err := scanCode(s.DB.QueryRow(`SELECT id, code, durationMonths, status, createdAt, usedAt, usedByUserId FROM redeemCodes WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

// Redeem consumes a code and extends membership atomically.
func (s *Store) Redeem(codeID, userID, phone, beforeExpire, afterExpire string) error {
	return s.inTx(func(tx *sql.Tx) error {
		now := Now()
		res, err := tx.Exec(`UPDATE redeemCodes SET status = 'used', usedAt = ?, usedByUserId = ? WHERE id = ? AND status = 'unused'`, now, userID, codeID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return errors.New("code already used")
		}
		if _, err := tx.Exec(`UPDATE users SET memberExpireAt = ? WHERE id = ?`, afterExpire, userID); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO redeemRecords (id, userId, phone, code, activatedAt, beforeExpireAt, afterExpireAt) VALUES (?,?,?,?,?,?,?)`,
			NewID(), userID, phone, "", now, nilIfEmpty(beforeExpire), afterExpire); err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) InsertRedeemCode(c *model.RedeemCode) error {
	_, err := s.DB.Exec(`INSERT INTO redeemCodes (id, code, durationMonths, status, createdAt, usedAt, usedByUserId) VALUES (?,?,?,?,?,?,?)`,
		c.ID, c.Code, c.DurationMonths, c.Status, c.CreatedAt, nil, nil)
	return err
}

func (s *Store) VoidRedeemCode(id string) error {
	res, err := s.DB.Exec(`UPDATE redeemCodes SET status = 'void' WHERE id = ? AND status = 'unused'`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListRedeemCodes(page, pageSize int, status, codeKeyword string) ([]model.RedeemCode, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}
	if codeKeyword != "" {
		where += " AND upper(code) LIKE ?"
		args = append(args, "%"+codeKeyword+"%")
	}
	var total int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM redeemCodes `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.DB.Query(`SELECT id, code, durationMonths, status, createdAt, usedAt, usedByUserId FROM redeemCodes `+where+` ORDER BY createdAt DESC LIMIT ? OFFSET ?`,
		append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.RedeemCode{}
	for rows.Next() {
		c, err := scanCode(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := s.attachPhones(out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *Store) attachPhones(codes []model.RedeemCode) error {
	for i := range codes {
		if codes[i].UsedByUserID == "" {
			continue
		}
		var phone string
		if err := s.DB.QueryRow(`SELECT phone FROM users WHERE id = ?`, codes[i].UsedByUserID).Scan(&phone); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return err
		}
		codes[i].UsedByPhone = phone
	}
	return nil
}

func (s *Store) ListRedeemCodesExport(status, codeKeyword string) ([]model.RedeemCode, error) {
	codes, _, err := s.ListRedeemCodes(1, 1000000, status, codeKeyword)
	return codes, err
}

// ---------- redeem records ----------

func (s *Store) InsertRedeemRecord(r *model.RedeemRecord) error {
	_, err := s.DB.Exec(`INSERT INTO redeemRecords (id, userId, phone, code, activatedAt, beforeExpireAt, afterExpireAt) VALUES (?,?,?,?,?,?,?)`,
		r.ID, r.UserID, r.Phone, r.Code, r.ActivatedAt, nilIfEmpty(r.BeforeExpireAt), r.AfterExpireAt)
	return err
}

func (s *Store) CountRedeemRecordsOnDay(dayStart, dayEnd string) (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM redeemRecords WHERE activatedAt >= ? AND activatedAt < ?`, dayStart, dayEnd).Scan(&n)
	return n, err
}

func (s *Store) ListRedeemRecords(page, pageSize int, phone, start, end string) ([]model.RedeemRecord, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	if phone != "" {
		where += " AND phone LIKE ?"
		args = append(args, "%"+phone+"%")
	}
	if start != "" {
		where += " AND activatedAt >= ?"
		args = append(args, start)
	}
	if end != "" {
		where += " AND activatedAt <= ?"
		args = append(args, end)
	}
	var total int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM redeemRecords `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.DB.Query(`SELECT id, userId, phone, code, activatedAt, beforeExpireAt, afterExpireAt FROM redeemRecords `+where+` ORDER BY activatedAt DESC LIMIT ? OFFSET ?`,
		append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.RedeemRecord{}
	for rows.Next() {
		var r model.RedeemRecord
		var before, after sql.NullString
		if err := rows.Scan(&r.ID, &r.UserID, &r.Phone, &r.Code, &r.ActivatedAt, &before, &after); err != nil {
			return nil, 0, err
		}
		r.BeforeExpireAt = before.String
		r.AfterExpireAt = after.String
		out = append(out, r)
	}
	return out, total, rows.Err()
}

// ---------- conversations (admin) ----------

func (s *Store) ListConversationsAdmin(page, pageSize int, phone, keyword string) ([]model.Conversation, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	if phone != "" {
		where += " AND userId IN (SELECT id FROM users WHERE phone LIKE ?)"
		args = append(args, "%"+phone+"%")
	}
	if keyword != "" {
		where += " AND title LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	var total int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM conversations `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.DB.Query(`SELECT `+convColumns+` FROM conversations `+where+` ORDER BY updatedAt DESC LIMIT ? OFFSET ?`,
		append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.Conversation{}
	for rows.Next() {
		c, err := scanConversation(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	for i := range out {
		var phone string
		if err := s.DB.QueryRow(`SELECT phone FROM users WHERE id = ?`, out[i].UserID).Scan(&phone); err == nil {
			out[i].UserPhone = phone
		}
	}
	return out, total, nil
}

func (s *Store) CountConversations() (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM conversations`).Scan(&n)
	return n, err
}

var _ = fmt.Sprintf
