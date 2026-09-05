package model

// Times are stored as ISO-8601 UTC strings identical to the Node backend
// ("2006-01-02T15:04:05.000Z") so one SQLite file serves both implementations.
// JSON tags mirror the exact field names the Vue frontend consumes.

type User struct {
	ID              string `json:"id"`
	Phone           string `json:"phone"`
	Nickname        string `json:"nickname"`
	PasswordHash    string `json:"-"`
	AvatarURL       string `json:"avatarUrl"`
	Status          string `json:"status"`
	Role            string `json:"role"`
	CurrentSession  string `json:"currentSessionId"`
	SessionUpdateAt string `json:"sessionUpdatedAt"`
	MemberExpireAt  string `json:"memberExpireAt"`
	CreatedAt       string `json:"createdAt"`
	LastLoginAt     string `json:"lastLoginAt"`
	AdminUsername   string `json:"adminUsername"`
	AnswerLength    string `json:"answerLength"`
	AnswerStyle     string `json:"answerStyle"`
}

type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Menu struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	MenuGroup string `json:"group"`
}

type RedeemCode struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	DurationMonths int    `json:"durationMonths"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt"`
	UsedAt         string `json:"usedAt"`
	UsedByUserID   string `json:"usedByUserId"`
	UsedByPhone    string `json:"usedByPhone,omitempty"`
}

type RedeemRecord struct {
	ID             string `json:"id"`
	UserID         string `json:"userId"`
	Phone          string `json:"phone"`
	Code           string `json:"code"`
	ActivatedAt    string `json:"activatedAt"`
	BeforeExpireAt string `json:"beforeExpireAt"`
	AfterExpireAt  string `json:"afterExpireAt"`
}

type Conversation struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	UserPhone string `json:"userPhone,omitempty"`
}

type Message struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversationId"`
	Role           string `json:"role"`
	Content        string `json:"content"`
	CreatedAt      string `json:"createdAt"`
}

type Announcement struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type AnnouncementRead struct {
	ID             string `json:"id"`
	AnnouncementID string `json:"announcementId"`
	UserID         string `json:"userId"`
	CreatedAt      string `json:"createdAt"`
}

type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SafeUser is the projection returned by user-facing endpoints (never leaks hashes).
func (u *User) SafeUser() map[string]any {
	return map[string]any{
		"id":             u.ID,
		"phone":          u.Phone,
		"nickname":       u.Nickname,
		"avatarUrl":      u.AvatarURL,
		"status":         u.Status,
		"role":           u.Role,
		"memberExpireAt": nilOrNil(u.MemberExpireAt),
		"createdAt":      u.CreatedAt,
		"answerLength":   nilOrNil(u.AnswerLength),
		"answerStyle":    nilOrNil(u.AnswerStyle),
	}
}

func nilOrNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}
