package store

import (
	"database/sql"
	"errors"
	"strings"

	"ai-chat-backend/internal/model"
)

// ---------- users ----------

const userColumns = `id, phone, nickname, passwordHash, avatarUrl, status, role,
	currentSessionId, sessionUpdatedAt, memberExpireAt, createdAt, lastLoginAt,
	adminUsername, answerLength, answerStyle`

func scanUser(row interface{ Scan(...any) error }) (*model.User, error) {
	u := &model.User{}
	var session, sessionUp, memberExpire, adminUser, answerLen, answerStyle sql.NullString
	err := row.Scan(&u.ID, &u.Phone, &u.Nickname, &u.PasswordHash, &u.AvatarURL, &u.Status,
		&u.Role, &session, &sessionUp, &memberExpire, &u.CreatedAt, &u.LastLoginAt,
		&adminUser, &answerLen, &answerStyle)
	if err != nil {
		return nil, err
	}
	u.CurrentSession = session.String
	u.SessionUpdateAt = sessionUp.String
	u.MemberExpireAt = memberExpire.String
	u.AdminUsername = adminUser.String
	u.AnswerLength = answerLen.String
	u.AnswerStyle = answerStyle.String
	return u, nil
}

func (s *Store) GetUserByID(id string) (*model.User, error) {
	u, err := scanUser(s.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (s *Store) GetUserByPhone(phone string) (*model.User, error) {
	u, err := scanUser(s.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE phone = ?`, phone))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (s *Store) GetAdminByUsername(username string) (*model.User, error) {
	u, err := scanUser(s.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE role = 'admin' AND adminUsername = ?`, username))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// IsUniqueViolation reports whether err is a SQLite UNIQUE constraint failure
// (used to turn concurrent duplicate inserts into a clean 400).
func IsUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func (s *Store) CreateUser(u *model.User) error {
	_, err := s.DB.Exec(`INSERT INTO users (id, phone, nickname, passwordHash, avatarUrl, status, role,
		currentSessionId, sessionUpdatedAt, memberExpireAt, createdAt, lastLoginAt, adminUsername, answerLength, answerStyle)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		u.ID, u.Phone, u.Nickname, u.PasswordHash, u.AvatarURL, u.Status, u.Role,
		nilIfEmpty(u.CurrentSession), nilIfEmpty(u.SessionUpdateAt), nilIfEmpty(u.MemberExpireAt),
		u.CreatedAt, u.LastLoginAt, nilIfEmpty(u.AdminUsername), nilIfEmpty(u.AnswerLength), nilIfEmpty(u.AnswerStyle))
	return err
}

func (s *Store) UpdateUserSession(id, sessionID string) error {
	_, err := s.DB.Exec(`UPDATE users SET currentSessionId = ?, sessionUpdatedAt = ? WHERE id = ?`,
		nilIfEmpty(sessionID), nilIfEmpty(Now()), id)
	return err
}

func (s *Store) UpdateUserLogin(id, sessionID string) error {
	return s.inTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`UPDATE users SET lastLoginAt = ?, currentSessionId = ?, sessionUpdatedAt = ? WHERE id = ?`,
			Now(), sessionID, Now(), id); err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) UpdateUserProfile(id, nickname, avatarURL string) error {
	_, err := s.DB.Exec(`UPDATE users SET nickname = ?, avatarUrl = ? WHERE id = ?`, nickname, avatarURL, id)
	return err
}

func (s *Store) UpdateUserPassword(id, hash string) error {
	_, err := s.DB.Exec(`UPDATE users SET passwordHash = ? WHERE id = ?`, hash, id)
	return err
}

func (s *Store) UpdateUserPreferences(id, length, style string) error {
	_, err := s.DB.Exec(`UPDATE users SET answerLength = ?, answerStyle = ? WHERE id = ?`,
		nilIfEmpty(length), nilIfEmpty(style), id)
	return err
}

func (s *Store) UpdateUserMemberExpire(id, expireAt string) error {
	_, err := s.DB.Exec(`UPDATE users SET memberExpireAt = ? WHERE id = ?`, nilIfEmpty(expireAt), id)
	return err
}

func (s *Store) UpdateUserStatus(id, status string) error {
	_, err := s.DB.Exec(`UPDATE users SET status = ? WHERE id = ?`, status, id)
	return err
}

func (s *Store) UpdateUserBase(id, phone, nickname string) error {
	_, err := s.DB.Exec(`UPDATE users SET phone = ?, nickname = ? WHERE id = ?`, phone, nickname, id)
	return err
}

func (s *Store) PhoneExists(phone string) (bool, error) {
	var one int
	err := s.DB.QueryRow(`SELECT 1 FROM users WHERE phone = ? LIMIT 1`, phone).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// ListUsers returns role='user' rows ordered by createdAt DESC, paged.
func (s *Store) ListUsers(page, pageSize int, phone, status string) ([]model.User, int, error) {
	where := `WHERE role = 'user'`
	args := []any{}
	if phone != "" {
		where += ` AND phone LIKE ? ESCAPE '\'`
		args = append(args, "%"+escapeLike(phone)+"%")
	}
	if status != "" {
		where += ` AND status = ?`
		args = append(args, status)
	}
	var total int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM users `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.DB.Query(`SELECT `+userColumns+` FROM users `+where+` ORDER BY createdAt DESC LIMIT ? OFFSET ?`,
		append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *u)
	}
	return out, total, rows.Err()
}

// MembersList returns every user with a non-empty memberExpireAt.
func (s *Store) MembersList() ([]model.User, error) {
	rows, err := s.DB.Query(`SELECT ` + userColumns + ` FROM users WHERE role = 'user' AND memberExpireAt IS NOT NULL AND memberExpireAt != '' ORDER BY createdAt DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func (s *Store) CountUsers(role string) (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = ?`, role).Scan(&n)
	return n, err
}

// CountActiveMembers counts users whose membership is still valid at call time.
func (s *Store) CountActiveMembers(nowISO string) (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'user' AND memberExpireAt IS NOT NULL AND memberExpireAt != '' AND memberExpireAt > ?`, nowISO).Scan(&n)
	return n, err
}

func (s *Store) GetRoles() ([]model.Role, error) {
	rows, err := s.DB.Query(`SELECT id, name, description FROM roles`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Role{}
	for rows.Next() {
		var r model.Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) GetMenus() ([]model.Menu, error) {
	rows, err := s.DB.Query(`SELECT id, name, path, menuGroup FROM menus`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Menu{}
	for rows.Next() {
		var m model.Menu
		if err := rows.Scan(&m.ID, &m.Name, &m.Path, &m.MenuGroup); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func nilIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
