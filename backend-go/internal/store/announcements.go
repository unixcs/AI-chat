package store

import (
	"strings"
	"database/sql"
	"errors"

	"ai-chat-backend/internal/model"
)

// ---------- announcements ----------

func scanAnnouncement(row interface{ Scan(...any) error }) (*model.Announcement, error) {
	a := &model.Announcement{}
	var active int
	err := row.Scan(&a.ID, &a.Title, &a.Content, &active, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	a.Active = active == 1
	return a, nil
}

const annColumns = `id, title, content, active, createdAt, updatedAt`

func (s *Store) CreateAnnouncement(a *model.Announcement) error {
	_, err := s.DB.Exec(`INSERT INTO announcements (id, title, content, active, createdAt, updatedAt) VALUES (?,?,?,?,?,?)`,
		a.ID, a.Title, a.Content, boolInt(a.Active), a.CreatedAt, a.UpdatedAt)
	return err
}

func (s *Store) UpdateAnnouncement(a *model.Announcement) error {
	res, err := s.DB.Exec(`UPDATE announcements SET title = ?, content = ?, active = ?, updatedAt = ? WHERE id = ?`,
		a.Title, a.Content, boolInt(a.Active), Now(), a.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteAnnouncement(id string) error {
	return s.inTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`DELETE FROM announcementReads WHERE announcementId = ?`, id); err != nil {
			return err
		}
		res, err := tx.Exec(`DELETE FROM announcements WHERE id = ?`, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *Store) GetAnnouncement(id string) (*model.Announcement, error) {
	a, err := scanAnnouncement(s.DB.QueryRow(`SELECT `+annColumns+` FROM announcements WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (s *Store) ListAnnouncements() ([]model.Announcement, error) {
	rows, err := s.DB.Query(`SELECT ` + annColumns + ` FROM announcements ORDER BY createdAt DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Announcement{}
	for rows.Next() {
		a, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

// CurrentUnreadAnnouncement returns the newest active announcement the user
// has not acknowledged yet, or ErrNotFound.
func (s *Store) CurrentUnreadAnnouncement(userID string) (*model.Announcement, error) {
	a, err := scanAnnouncement(s.DB.QueryRow(`SELECT `+annColumns+` FROM announcements
		WHERE active = 1 AND id NOT IN (SELECT announcementId FROM announcementReads WHERE userId = ?)
		ORDER BY createdAt DESC LIMIT 1`, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (s *Store) AckAnnouncement(announcementID, userID string) error {
	_, err := s.DB.Exec(`INSERT OR IGNORE INTO announcementReads (id, announcementId, userId, createdAt) VALUES (?,?,?,?)`,
		NewID(), announcementID, userID, Now())
	return err
}

func (s *Store) CountAnnouncementReads(announcementID string) (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM announcementReads WHERE announcementId = ?`, announcementID).Scan(&n)
	return n, err
}

// ---------- settings ----------

func (s *Store) GetSetting(key, fallback string) string {
	var v string
	err := s.DB.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	if err != nil {
		return fallback
	}
	return v
}

func (s *Store) SetSetting(key, value string) error {
	_, err := s.DB.Exec(`INSERT INTO settings (key, value) VALUES (?,?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// escapeLike neutralizes LIKE wildcards so filters match literally,
// mirroring the Node backend's String.includes semantics.
func escapeLike(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, "%", `\%`)
	v = strings.ReplaceAll(v, "_", `\_`)
	return v
}
