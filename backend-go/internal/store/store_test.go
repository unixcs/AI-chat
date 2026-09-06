package store

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.DB.Close() })
	return s
}

func TestSeedAdmin(t *testing.T) {
	s := openTest(t)
	admin, err := s.GetAdminByUsername("admin")
	if err != nil {
		t.Fatalf("seeded admin missing: %v", err)
	}
	if !CheckPassword(admin.PasswordHash, "admin123") {
		t.Fatal("seeded admin password mismatch")
	}
	menus, err := s.GetMenus()
	if err != nil || len(menus) < 8 {
		t.Fatalf("menus not seeded: %v %d", err, len(menus))
	}
}

func TestUserLifecycle(t *testing.T) {
	s := openTest(t)
	hash, _ := HashPassword("secret1")
	if err := s.CreateUser(userFixture("u1", "13800000001", hash)); err != nil {
		t.Fatalf("create user: %v", err)
	}
	got, err := s.GetUserByPhone("13800000001")
	if err != nil {
		t.Fatalf("get by phone: %v", err)
	}
	if got.ID != "u1" {
		t.Fatalf("unexpected id %s", got.ID)
	}
	if err := s.UpdateUserLogin("u1", "sess_abc"); err != nil {
		t.Fatalf("login update: %v", err)
	}
	after, _ := s.GetUserByID("u1")
	if after.CurrentSession != "sess_abc" {
		t.Fatalf("session not stored: %q", after.CurrentSession)
	}
	if err := s.UpdateUserSession("u1", ""); err != nil {
		t.Fatalf("logout: %v", err)
	}
	after2, _ := s.GetUserByID("u1")
	if after2.CurrentSession != "" {
		t.Fatalf("session not cleared: %q", after2.CurrentSession)
	}
}

func TestRedeemFlow(t *testing.T) {
	s := openTest(t)
	hash, _ := HashPassword("secret1")
	_ = s.CreateUser(userFixture("u1", "13800000001", hash))
	code := redeemFixture("c1", "VIP-TEST-000001", 1)
	if err := s.InsertRedeemCode(code); err != nil {
		t.Fatalf("insert code: %v", err)
	}

	got, err := s.GetRedeemCodeByCode("VIP-TEST-000001")
	if err != nil || got.Status != "unused" {
		t.Fatalf("code lookup: %v %+v", err, got)
	}

	after := AddMonths(Now(), 1)
	if rerr := s.Redeem("c1", "VIP-TEST-000001", "u1", "13800000001", 1); rerr != nil {
		t.Fatalf("redeem: %v", rerr)
	}
	user, _ := s.GetUserByID("u1")
	if user.MemberExpireAt != after {
		t.Fatalf("membership not extended: %q", user.MemberExpireAt)
	}
	records, _, err := s.ListRedeemRecords(1, 10, "", "", "")
	if err != nil || len(records) != 1 {
		t.Fatalf("records: %v %d", err, len(records))
	}
	if records[0].Code != "VIP-TEST-000001" {
		t.Fatalf("redeem record must carry the code, got %q", records[0].Code)
	}
	// second redeem of same code must fail
	if rerr := s.Redeem("c1", "VIP-TEST-000001", "u1", "13800000001", 1); rerr == nil {
		t.Fatal("double redeem must fail")
	}
}

func TestConversationDeleteCascades(t *testing.T) {
	s := openTest(t)
	if err := s.CreateConversation(convFixture("conv1", "u1")); err != nil {
		t.Fatalf("create conv: %v", err)
	}
	_ = s.InsertMessage(msgFixture("m1", "conv1", "user", "hello"))
	_ = s.InsertMessage(msgFixture("m2", "conv1", "assistant", "hi"))
	if err := s.DeleteConversation("conv1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	msgs, _ := s.GetMessages("conv1")
	if len(msgs) != 0 {
		t.Fatalf("messages must cascade-delete, got %d", len(msgs))
	}
	if err := s.DeleteConversation("conv1"); err != ErrNotFound {
		t.Fatalf("second delete must be ErrNotFound, got %v", err)
	}
}

func TestAnnouncementUnreadFlow(t *testing.T) {
	s := openTest(t)
	ann := annFixture("a1", "维护通知", "今晚维护", true)
	if err := s.CreateAnnouncement(ann); err != nil {
		t.Fatalf("create ann: %v", err)
	}
	got, err := s.CurrentUnreadAnnouncement("u1")
	if err != nil || got.ID != "a1" {
		t.Fatalf("unread ann: %v %+v", err, got)
	}
	if err := s.AckAnnouncement("a1", "u1"); err != nil {
		t.Fatalf("ack: %v", err)
	}
	if _, err := s.CurrentUnreadAnnouncement("u1"); err != ErrNotFound {
		t.Fatalf("acked ann must be unread no more, got %v", err)
	}
	// other users still see it
	if _, err := s.CurrentUnreadAnnouncement("u2"); err != nil {
		t.Fatalf("other user should still see it: %v", err)
	}
}

func TestUpdateAnnouncementResetReads(t *testing.T) {
	s := openTest(t)
	ann := annFixture("a1", "维护通知", "今晚维护", true)
	if err := s.CreateAnnouncement(ann); err != nil {
		t.Fatalf("create ann: %v", err)
	}
	if err := s.AckAnnouncement("a1", "u1"); err != nil {
		t.Fatalf("ack: %v", err)
	}
	ann.Content = "改期到明晚维护"
	if err := s.UpdateAnnouncementResetReads(ann); err != nil {
		t.Fatalf("update+reset: %v", err)
	}
	got, err := s.CurrentUnreadAnnouncement("u1")
	if err != nil || got.ID != "a1" || got.Content != "改期到明晚维护" {
		t.Fatalf("reset must re-notify: %v %+v", err, got)
	}
	if err := s.UpdateAnnouncementResetReads(annFixture("missing", "x", "y", true)); err != ErrNotFound {
		t.Fatalf("missing ann must be ErrNotFound, got %v", err)
	}
}

func TestSettingsUpsert(t *testing.T) {
	s := openTest(t)
	if v := s.GetSetting("registrationOpen", "x"); v != "x" {
		t.Fatalf("fallback broken: %q", v)
	}
	_ = s.SetSetting("registrationOpen", "false")
	if v := s.GetSetting("registrationOpen", "x"); v != "false" {
		t.Fatalf("set broken: %q", v)
	}
	_ = s.SetSetting("registrationOpen", "true")
	if v := s.GetSetting("registrationOpen", "x"); v != "true" {
		t.Fatalf("upsert broken: %q", v)
	}
}

func TestAddMonthsMatchesNodeSemantics(t *testing.T) {
	// Node: new Date('2026-01-31T00:00:00.000Z'), setMonth(+1) → Mar 3.
	got := AddMonths("2026-01-31T00:00:00.000Z", 1)
	if got != "2026-03-03T00:00:00.000Z" {
		t.Fatalf("AddMonths mismatch with Node dayjs/setMonth semantics: %s", got)
	}
	got2 := AddMonths("2026-01-15T00:00:00.000Z", 1)
	if got2 != "2026-02-15T00:00:00.000Z" {
		t.Fatalf("AddMonths plain case: %s", got2)
	}
}

// Concurrency CAS: exactly one of N concurrent redeems of the same code wins.
func TestRedeemConcurrentDoubleSpend(t *testing.T) {
	s := openTest(t)
	hash, _ := HashPassword("secret1")
	_ = s.CreateUser(userFixture("u1", "13800000001", hash))
	_ = s.InsertRedeemCode(redeemFixture("c1", "VIP-TEST-CONCURRENT", 1))

	const n = 8
	winners := make(chan struct{}, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Redeem("c1", "VIP-TEST-CONCURRENT", "u1", "13800000001", 1); err == nil {
				winners <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(winners)
	count := 0
	for range winners {
		count++
	}
	if count != 1 {
		t.Fatalf("exactly one concurrent redeem must win, got %d", count)
	}
	user, _ := s.GetUserByID("u1")
	if user.MemberExpireAt == "" {
		t.Fatal("winner must extend membership")
	}
}

// P0-2 regression: concurrent redeems of DIFFERENT codes by the same user must
// stack (serialized transactions), not lose months to read-modify-write races.
func TestRedeemConcurrentStacking(t *testing.T) {
	s := openTest(t)
	hash, _ := HashPassword("secret1")
	_ = s.CreateUser(userFixture("u1", "13800000001", hash))
	const n = 5
	for i := 0; i < n; i++ {
		_ = s.InsertRedeemCode(redeemFixture(fmt.Sprintf("c%d", i), fmt.Sprintf("VIP-STACK-%04d", i), 1))
	}
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := s.Redeem(fmt.Sprintf("c%d", i), fmt.Sprintf("VIP-STACK-%04d", i), "u1", "13800000001", 1); err != nil {
				t.Errorf("redeem %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()
	user, _ := s.GetUserByID("u1")
	elapsed := TimeValue(user.MemberExpireAt).Sub(time.Now())
	days := elapsed.Hours() / 24
	if days < 145 || days > 160 { // ~5 months
		t.Fatalf("5 concurrent months must stack, got %.1f days (%s)", days, user.MemberExpireAt)
	}
}
