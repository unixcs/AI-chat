// Package store owns all SQLite access. Unlike the previous Node backend —
// which re-wrote every table on each mutation — every operation here is a
// targeted statement inside a transaction where needed.
package store

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"ai-chat-backend/internal/model"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	DB *sql.DB
}

func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(0)&_pragma=synchronous(NORMAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite is a single-writer database; a small pool avoids SQLITE_BUSY storms.
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	db.SetConnMaxLifetime(0)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{DB: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	if err := s.seed(); err != nil {
		return nil, err
	}
	s.sweepOrphans()
	return s, nil
}

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY, phone TEXT, nickname TEXT, passwordHash TEXT,
			avatarUrl TEXT, status TEXT, role TEXT, currentSessionId TEXT,
			sessionUpdatedAt TEXT, memberExpireAt TEXT, createdAt TEXT,
			lastLoginAt TEXT, adminUsername TEXT);`,
		`CREATE TABLE IF NOT EXISTS roles (id TEXT PRIMARY KEY, name TEXT, description TEXT);`,
		`CREATE TABLE IF NOT EXISTS menus (id TEXT PRIMARY KEY, name TEXT, path TEXT, menuGroup TEXT);`,
		`CREATE TABLE IF NOT EXISTS redeemCodes (
			id TEXT PRIMARY KEY, code TEXT, durationMonths INTEGER, status TEXT,
			createdAt TEXT, usedAt TEXT, usedByUserId TEXT);`,
		`CREATE TABLE IF NOT EXISTS redeemRecords (
			id TEXT PRIMARY KEY, userId TEXT, phone TEXT, code TEXT,
			activatedAt TEXT, beforeExpireAt TEXT, afterExpireAt TEXT);`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY, userId TEXT, title TEXT, createdAt TEXT, updatedAt TEXT);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY, conversationId TEXT, role TEXT, content TEXT, createdAt TEXT);`,
		`CREATE TABLE IF NOT EXISTS announcements (
			id TEXT PRIMARY KEY, title TEXT, content TEXT, active INTEGER,
			createdAt TEXT, updatedAt TEXT);`,
		`CREATE TABLE IF NOT EXISTS announcementReads (
			id TEXT PRIMARY KEY, announcementId TEXT, userId TEXT, createdAt TEXT);`,
		`CREATE TABLE IF NOT EXISTS promptRevisions (
			id TEXT PRIMARY KEY, version INTEGER UNIQUE, content TEXT NOT NULL,
			operator TEXT NOT NULL, createdAt TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversationId);`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_user ON conversations(userId);`,
		`CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);`,
		// Best-effort hard constraint: fails silently on legacy DBs that
		// already contain duplicate phones (those keep the old behavior).
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique ON users(phone);`,
		`CREATE INDEX IF NOT EXISTS idx_read_ann_user ON announcementReads(announcementId, userId);`,
	}
	for _, stmt := range stmts {
		if _, err := s.DB.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return s.addUsersColumns()
}

// addUsersColumns adds new nullable columns non-destructively (old DBs keep working).
func (s *Store) addUsersColumns() error {
	rows, err := s.DB.Query(`PRAGMA table_info(users)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	existing := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			return err
		}
		existing[name] = true
	}
	rows.Close()

	for _, col := range []string{"answerLength", "answerStyle"} {
		if !existing[col] {
			if _, err := s.DB.Exec("ALTER TABLE users ADD COLUMN " + col + " TEXT"); err != nil {
				return fmt.Errorf("add column %s: %w", col, err)
			}
		}
	}
	return nil
}

// sweepOrphans removes messages whose conversation no longer exists (they can
// only originate from legacy data or crashes; the live path now drops them).
func (s *Store) sweepOrphans() {
	_, _ = s.DB.Exec(`DELETE FROM messages WHERE conversationId NOT IN (SELECT id FROM conversations)`)
}

func (s *Store) seed() error {
	var count int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return s.ensureMenus()
	}
	return s.inTx(func(tx *sql.Tx) error {
		now := Now()
		hash, err := HashPassword("admin123")
		if err != nil {
			return err
		}
		adminID := NewID()
		if _, err := tx.Exec(`INSERT INTO users (id, phone, nickname, passwordHash, avatarUrl, status, role,
			currentSessionId, sessionUpdatedAt, memberExpireAt, createdAt, lastLoginAt, adminUsername)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			adminID, "10000000000", "系统管理员", hash, "", "active", "admin",
			nil, nil, nil, now, now, "admin"); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO roles (id, name, description) VALUES ('role_admin','admin','系统管理员'),('role_user','user','普通用户')`); err != nil {
			return err
		}
		menus := defaultMenus()
		for _, m := range menus {
			if _, err := tx.Exec(`INSERT INTO menus (id, name, path, menuGroup) VALUES (?,?,?,?)`, m.ID, m.Name, m.Path, m.MenuGroup); err != nil {
				return err
			}
		}
		return nil
	})
}

// ensureMenus keeps new admin pages (announcements / AI status) visible on DBs
// created by the Node backend, without touching any other row.
func (s *Store) ensureMenus() error {
	for _, m := range defaultMenus() {
		var one int
		if err := s.DB.QueryRow(`SELECT 1 FROM menus WHERE path = ?`, m.Path).Scan(&one); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if _, err := s.DB.Exec(`INSERT INTO menus (id, name, path, menuGroup) VALUES (?,?,?,?)`, m.ID, m.Name, m.Path, m.MenuGroup); err != nil {
					return err
				}
				continue
			}
			return err
		}
	}
	return nil
}

func defaultMenus() []model.Menu {
	now := Now()
	return []model.Menu{
		{ID: NewID(), Name: "控制台", Path: "/admin", MenuGroup: "首页"},
		{ID: NewID(), Name: "用户管理", Path: "/admin/users", MenuGroup: "系统管理"},
		{ID: NewID(), Name: "角色管理", Path: "/admin/roles", MenuGroup: "系统管理"},
		{ID: NewID(), Name: "菜单管理", Path: "/admin/menus", MenuGroup: "系统管理"},
		{ID: NewID(), Name: "会员管理", Path: "/admin/members", MenuGroup: "业务管理"},
		{ID: NewID(), Name: "兑换码管理", Path: "/admin/redeem-codes", MenuGroup: "业务管理"},
		{ID: NewID(), Name: "兑换记录", Path: "/admin/redeem-records", MenuGroup: "业务管理"},
		{ID: NewID(), Name: "会话管理", Path: "/admin/conversations", MenuGroup: "业务管理"},
		{ID: "menu_announcements_" + now, Name: "公告管理", Path: "/admin/announcements", MenuGroup: "业务管理"},
		{ID: "menu_aistatus_" + now, Name: "AI 状态", Path: "/admin/ai", MenuGroup: "系统管理"},
		{ID: "menu_prompt_" + now, Name: "提示词管理", Path: "/admin/prompt", MenuGroup: "系统管理"},
	}
}

func (s *Store) inTx(fn func(tx *sql.Tx) error) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// promptVersionMu serializes prompt-version allocation. The transaction is
// DEFERRED, so two concurrent read-MAX-then-INSERT transactions on different
// pool connections could both see the same snapshot (the loser would hit
// SQLITE_BUSY_SNAPSHOT, which busy_timeout does not retry). Single-process
// deployment makes an in-process write mutex equivalent to BEGIN IMMEDIATE
// and keeps the tx helpers untouched.
var promptVersionMu sync.Mutex

// ---------- ids & time (Node-compatible formats) ----------

const idChars = "abcdefghijklmnopqrstuvwxyz0123456789"

func NewID() string {
	return fmt.Sprintf("%d_%s", time.Now().UnixMilli(), randString(6))
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(idChars))))
		b[i] = idChars[idx.Int64()]
	}
	return string(b)
}

func randUpper(n int) string {
	return strings.ToUpper(randString(n))
}

func Now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
}

func AddMonths(value string, months int) string {
	t, err := time.Parse(time.RFC3339, normalizeISO(value))
	if err != nil {
		return value
	}
	t = t.AddDate(0, months, 0)
	return t.UTC().Format("2006-01-02T15:04:05.000Z07:00")
}

// normalizeISO tolerates the several ISO shapes stored by dayjs/Node.
func normalizeISO(v string) string {
	if v == "" {
		return v
	}
	if strings.HasSuffix(v, "Z") || strings.Contains(v, "+") || strings.Contains(v[10:], "-") {
		if len(v) == 16 { // "2006-01-02T15:04Z"
			return v
		}
	}
	t, err := time.Parse("2006-01-02T15:04:05.000Z07:00", v)
	if err == nil {
		return t.Format(time.RFC3339Nano)
	}
	t, err = time.Parse("2006-01-02T15:04:05Z07:00", v)
	if err == nil {
		return t.Format(time.RFC3339Nano)
	}
	return v
}

// TimeValue parses any stored ISO string; zero time when unparsable.
func TimeValue(v string) time.Time {
	for _, layout := range []string{"2006-01-02T15:04:05.000Z07:00", time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, v); err == nil {
			return t
		}
	}
	return time.Time{}
}
