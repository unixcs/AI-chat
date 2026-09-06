package store

import (
	"database/sql"
	"errors"

	"ai-chat-backend/internal/model"
)

// ---------- prompt revisions (append-only) ----------

// CreatePromptRevision allocates MAX(version)+1 under the write mutex and
// inserts the new head. Returns the allocated version.
func (s *Store) CreatePromptRevision(content, operator string) (int64, error) {
	promptVersionMu.Lock()
	defer promptVersionMu.Unlock()

	var version int64
	err := s.inTx(func(tx *sql.Tx) error {
		if err := tx.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM promptRevisions`).Scan(&version); err != nil {
			return err
		}
		version++
		_, err := tx.Exec(`INSERT INTO promptRevisions (id, version, content, operator, createdAt) VALUES (?,?,?,?,?)`,
			NewID(), version, content, operator, Now())
		return err
	})
	if err != nil {
		return 0, err
	}
	return version, nil
}

// CurrentPromptRevision returns the highest-version row (the effective prompt).
func (s *Store) CurrentPromptRevision() (*model.PromptRevision, error) {
	r, err := scanPromptRevision(s.DB.QueryRow(
		`SELECT id, version, content, operator, createdAt FROM promptRevisions ORDER BY version DESC LIMIT 1`))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return r, err
}

// GetPromptRevision fetches one row by version number.
func (s *Store) GetPromptRevision(version int64) (*model.PromptRevision, error) {
	r, err := scanPromptRevision(s.DB.QueryRow(
		`SELECT id, version, content, operator, createdAt FROM promptRevisions WHERE version = ?`, version))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return r, err
}

// ListPromptRevisions returns the newest `limit` rows, version DESC.
func (s *Store) ListPromptRevisions(limit int) ([]model.PromptRevision, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.DB.Query(
		`SELECT id, version, content, operator, createdAt FROM promptRevisions ORDER BY version DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.PromptRevision{}
	for rows.Next() {
		r, err := scanPromptRevision(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, rows.Err()
}

func scanPromptRevision(row interface{ Scan(...any) error }) (*model.PromptRevision, error) {
	r := &model.PromptRevision{}
	if err := row.Scan(&r.ID, &r.Version, &r.Content, &r.Operator, &r.CreatedAt); err != nil {
		return nil, err
	}
	return r, nil
}
