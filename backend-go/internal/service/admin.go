package service

import (
	"fmt"
	"strings"
	"time"

	"ai-chat-backend/internal/model"
	"ai-chat-backend/internal/store"
)

// PageResult is the generic paginated envelope the admin frontend expects.
type PageResult struct {
	Items    any   `json:"items"`
	Total    int   `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

func pageParams(page, pageSize, maxPageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func (s *Service) AdminLogin(username, password string) (string, *ServiceError) {
	admin, err := s.Store.GetAdminByUsername(username)
	if err != nil {
		return "", fail(400, "管理员账号不存在")
	}
	if !store.CheckPassword(admin.PasswordHash, password) {
		return "", fail(400, "管理员密码错误")
	}
	if serr := s.Store.UpdateUserLogin(admin.ID, ""); serr != nil {
		return "", fail(500, "服务器繁忙")
	}
	token, serr := adminToken(s.Cfg.JWTSecret, admin.ID)
	if serr != nil {
		return "", serr
	}
	return token, nil
}

func (s *Service) AdminChangePassword(adminID, newPassword string) *ServiceError {
	if len(newPassword) < 6 {
		return fail(400, "新密码至少 6 位")
	}
	admin, err := s.Store.GetUserByID(adminID)
	if err != nil || admin.Role != "admin" {
		return fail(404, "管理员不存在")
	}
	hash, err := store.HashPassword(newPassword)
	if err != nil {
		return fail(500, "服务器繁忙")
	}
	if err := s.Store.UpdateUserPassword(adminID, hash); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func (s *Service) Dashboard() (map[string]any, *ServiceError) {
	now := time.Now()
	userCount, err := s.Store.CountUsers("user")
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	memberCount, err := s.Store.CountActiveMembers(store.Now())
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	cst := now.In(time.FixedZone("CST", 8*3600))
	cstZone := cst.Location()
	dayStart := time.Date(cst.Year(), cst.Month(), cst.Day(), 0, 0, 0, 0, cstZone).UTC().Format("2006-01-02T15:04:05.000Z07:00")
	dayEnd := time.Date(cst.Year(), cst.Month(), cst.Day(), 0, 0, 0, 0, cstZone).Add(24 * time.Hour).UTC().Format("2006-01-02T15:04:05.000Z07:00")
	todayRedeemCount, err := s.Store.CountRedeemRecordsOnDay(dayStart, dayEnd)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	conversationCount, err := s.Store.CountConversations()
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return map[string]any{
		"userCount":        userCount,
		"memberCount":      memberCount,
		"todayRedeemCount": todayRedeemCount,
		"conversationCount": conversationCount,
	}, nil
}

func (s *Service) AdminUsers(page, pageSize int, phone, status string) (*PageResult, *ServiceError) {
	page, pageSize = pageParams(page, pageSize, 100)
	users, total, err := s.Store.ListUsers(page, pageSize, phone, status)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	items := make([]map[string]any, 0, len(users))
	for i := range users {
		u := &users[i]
		items = append(items, map[string]any{
			"id":            u.ID,
			"phone":         u.Phone,
			"nickname":      u.Nickname,
			"role":          u.Role,
			"status":        u.Status,
			"memberExpireAt": nilOrNil(u.MemberExpireAt),
			"createdAt":     u.CreatedAt,
		})
	}
	return &PageResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) AdminUpdateUser(id, nickname, phone string) *ServiceError {
	user, err := s.Store.GetUserByID(id)
	if err != nil || user.Role != "user" {
		return fail(404, "用户不存在")
	}
	if phone != "" && phone != user.Phone {
		if !phoneRe.MatchString(phone) {
			return fail(400, "手机号格式不正确")
		}
		exists, err := s.Store.PhoneExists(phone)
		if err != nil {
			return fail(500, "服务器繁忙")
		}
		if exists {
			return fail(400, "手机号已存在")
		}
	}
	nextNickname := user.Nickname
	if strings.TrimSpace(nickname) != "" {
		nextNickname = strings.TrimSpace(nickname)
	}
	nextPhone := user.Phone
	if phone != "" {
		nextPhone = phone
	}
	if err := s.Store.UpdateUserBase(id, nextPhone, nextNickname); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func (s *Service) AdminSetUserStatus(id, status string) *ServiceError {
	if status != "active" && status != "disabled" {
		return fail(400, "状态参数错误")
	}
	user, err := s.Store.GetUserByID(id)
	if err != nil || user.Role != "user" {
		return fail(404, "用户不存在")
	}
	if err := s.Store.UpdateUserStatus(id, status); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func (s *Service) AdminResetPassword(id, newPassword string) *ServiceError {
	if len(newPassword) < 6 {
		return fail(400, "新密码至少 6 位")
	}
	user, err := s.Store.GetUserByID(id)
	if err != nil || user.Role != "user" {
		return fail(404, "用户不存在")
	}
	hash, err := store.HashPassword(newPassword)
	if err != nil {
		return fail(500, "服务器繁忙")
	}
	if err := s.Store.UpdateUserPassword(id, hash); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

// NormalizeMemberExpireAt parses admin-provided dates like dayjs does:
// ISO strings, "YYYY-MM-DD HH:mm", or "YYYY-MM-DD". Empty → null (clears).
func NormalizeMemberExpireAt(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	layouts := []string{
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, value); err == nil {
			return t.UTC().Format("2006-01-02T15:04:05.000Z07:00"), nil
		}
	}
	return "", fmt.Errorf("会员到期时间格式不正确")
}

func (s *Service) AdminSetMemberExpire(id, memberExpireAt string) (*model.User, *ServiceError) {
	user, err := s.Store.GetUserByID(id)
	if err != nil || user.Role != "user" {
		return nil, fail(404, "用户不存在")
	}
	normalized, normErr := NormalizeMemberExpireAt(memberExpireAt)
	if normErr != nil {
		return nil, fail(400, "%s", normErr.Error())
	}
	if err := s.Store.UpdateUserMemberExpire(id, normalized); err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	updated, err := s.Store.GetUserByID(id)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return updated, nil
}

func (s *Service) AdminMembers() ([]map[string]any, *ServiceError) {
	users, err := s.Store.MembersList()
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	out := make([]map[string]any, 0, len(users))
	for i := range users {
		u := &users[i]
		out = append(out, map[string]any{
			"id":            u.ID,
			"phone":         u.Phone,
			"nickname":      u.Nickname,
			"memberExpireAt": u.MemberExpireAt,
		})
	}
	return out, nil
}

func (s *Service) AdminRedeemCodes(page, pageSize int, status, codeKeyword string) (*PageResult, *ServiceError) {
	page, pageSize = pageParams(page, pageSize, 100)
	codes, total, err := s.Store.ListRedeemCodes(page, pageSize, status, strings.ToUpper(codeKeyword))
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return &PageResult{Items: codes, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) AdminRedeemCodesExport(status, codeKeyword string) ([]model.RedeemCode, *ServiceError) {
	codes, err := s.Store.ListRedeemCodesExport(status, strings.ToUpper(codeKeyword))
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return codes, nil
}

// GenerateRedeemCodes mints up to 200 codes with the same VIP-YYYYMMDD-XXXXXX shape.
func (s *Service) GenerateRedeemCodes(quantity, durationMonths int) ([]model.RedeemCode, *ServiceError) {
	if quantity <= 0 {
		quantity = 1
	}
	if quantity > 200 {
		quantity = 200
	}
	if durationMonths <= 0 {
		durationMonths = 1
	}
	now := store.Now()
	created := make([]model.RedeemCode, 0, quantity)
	seen := map[string]bool{}
	for i := 0; i < quantity; i++ {
		var code string
		for {
			code = fmt.Sprintf("VIP-%s-%s", time.Now().UTC().Format("20060102"), randUpper(6))
			if !seen[code] {
				seen[code] = true
				break
			}
		}
		c := model.RedeemCode{
			ID:             store.NewID(),
			Code:           code,
			DurationMonths: durationMonths,
			Status:         "unused",
			CreatedAt:      now,
		}
		if err := s.Store.InsertRedeemCode(&c); err != nil {
			return nil, fail(500, "服务器繁忙")
		}
		created = append(created, c)
	}
	return created, nil
}

func (s *Service) AdminVoidRedeemCode(id string) *ServiceError {
	c, err := s.Store.GetRedeemCodeByID(id)
	if err != nil {
		return fail(404, "兑换码不存在")
	}
	if c.Status != "unused" {
		return fail(400, "仅未使用的兑换码可作废")
	}
	if err := s.Store.VoidRedeemCode(id); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func (s *Service) AdminRedeemRecords(page, pageSize int, phone, start, end string) (*PageResult, *ServiceError) {
	page, pageSize = pageParams(page, pageSize, 100)
	records, total, err := s.Store.ListRedeemRecords(page, pageSize, phone, start, end)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return &PageResult{Items: records, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) AdminConversations(page, pageSize int, phone, keyword string) (*PageResult, *ServiceError) {
	page, pageSize = pageParams(page, pageSize, 100)
	list, total, err := s.Store.ListConversationsAdmin(page, pageSize, phone, keyword)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return &PageResult{Items: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) AdminConversationMessages(conversationID string) ([]model.Message, *ServiceError) {
	msgs, err := s.Store.GetMessages(conversationID)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return msgs, nil
}

// SearchConversations filters by title/phone/content keyword (admin `search` param).
func (s *Service) SearchConversationIDs(keyword string) (map[string]bool, error) {
	return s.Store.SearchConversationIDsByContent(strings.ToLower(keyword))
}

// ---------- announcements (admin side) ----------

func (s *Service) AdminListAnnouncements() ([]map[string]any, *ServiceError) {
	list, err := s.Store.ListAnnouncements()
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	out := make([]map[string]any, 0, len(list))
	for i := range list {
		a := &list[i]
		reads, _ := s.Store.CountAnnouncementReads(a.ID)
		out = append(out, map[string]any{
			"id":        a.ID,
			"title":     a.Title,
			"content":   a.Content,
			"active":    a.Active,
			"createdAt": a.CreatedAt,
			"updatedAt": a.UpdatedAt,
			"readCount": reads,
		})
	}
	return out, nil
}

func (s *Service) AdminCreateAnnouncement(title, content string, active bool) (*model.Announcement, *ServiceError) {
	if strings.TrimSpace(title) == "" {
		return nil, fail(400, "公告标题不能为空")
	}
	if strings.TrimSpace(content) == "" {
		return nil, fail(400, "公告内容不能为空")
	}
	now := store.Now()
	a := &model.Announcement{
		ID:        store.NewID(),
		Title:     title,
		Content:   content,
		Active:    active,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.Store.CreateAnnouncement(a); err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return a, nil
}

func (s *Service) AdminUpdateAnnouncement(id, title, string_content string, active *bool) *ServiceError {
	a, err := s.Store.GetAnnouncement(id)
	if err != nil {
		return fail(404, "公告不存在")
	}
	if strings.TrimSpace(title) != "" {
		a.Title = title
	}
	if strings.TrimSpace(string_content) != "" {
		a.Content = string_content
	}
	if active != nil {
		a.Active = *active
	}
	if err := s.Store.UpdateAnnouncement(a); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func (s *Service) AdminDeleteAnnouncement(id string) *ServiceError {
	if err := s.Store.DeleteAnnouncement(id); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func nilOrNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}
