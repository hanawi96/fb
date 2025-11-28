package db

// ============================================
// NOTIFICATIONS METHODS
// ============================================

// GetUnreadNotifications lấy danh sách thông báo chưa đọc
func (s *Store) GetUnreadNotifications() ([]Notification, error) {
	query := `
		SELECT id, type, title, message, account_id, page_id, is_read, created_at
		FROM notifications
		WHERE is_read = false
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(
			&n.ID, &n.Type, &n.Title, &n.Message,
			&n.AccountID, &n.PageID, &n.IsRead, &n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// GetAllNotifications lấy tất cả thông báo (có phân trang)
func (s *Store) GetAllNotifications(limit, offset int) ([]Notification, error) {
	query := `
		SELECT id, type, title, message, account_id, page_id, is_read, created_at
		FROM notifications
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(
			&n.ID, &n.Type, &n.Title, &n.Message,
			&n.AccountID, &n.PageID, &n.IsRead, &n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// GetUnreadCount đếm số thông báo chưa đọc
func (s *Store) GetUnreadNotificationCount() (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM notifications WHERE is_read = false
	`).Scan(&count)
	return count, err
}

// MarkNotificationAsRead đánh dấu đã đọc
func (s *Store) MarkNotificationAsRead(id string) error {
	_, err := s.db.Exec(`
		UPDATE notifications SET is_read = true WHERE id = $1
	`, id)
	return err
}

// MarkAllNotificationsAsRead đánh dấu tất cả đã đọc
func (s *Store) MarkAllNotificationsAsRead() error {
	_, err := s.db.Exec(`UPDATE notifications SET is_read = true`)
	return err
}

// CreateNotification tạo thông báo mới
func (s *Store) CreateNotification(n *Notification) error {
	query := `
		INSERT INTO notifications (type, title, message, account_id, page_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	return s.db.QueryRow(
		query,
		n.Type, n.Title, n.Message, n.AccountID, n.PageID,
	).Scan(&n.ID, &n.CreatedAt)
}

// DeleteNotification xóa thông báo
func (s *Store) DeleteNotification(id string) error {
	_, err := s.db.Exec(`DELETE FROM notifications WHERE id = $1`, id)
	return err
}

// DeleteOldNotifications xóa thông báo cũ (> 30 ngày)
func (s *Store) DeleteOldNotifications() error {
	_, err := s.db.Exec(`
		DELETE FROM notifications 
		WHERE created_at < NOW() - INTERVAL '30 days'
	`)
	return err
}

// ============================================
// NOTIFICATION HELPERS
// ============================================

// NotifyRateLimit tạo thông báo rate limit
func (s *Store) NotifyRateLimit(accountID string, accountName string) error {
	n := &Notification{
		Type:      "rate_limit",
		Title:     "Nick bị rate limit",
		Message:   "Nick " + accountName + " đã bị Facebook rate limit. Vui lòng chờ 30 phút.",
		AccountID: &accountID,
	}
	return s.CreateNotification(n)
}

// NotifyDailyLimit tạo thông báo đạt giới hạn ngày
func (s *Store) NotifyDailyLimit(accountID string, accountName string) error {
	n := &Notification{
		Type:      "daily_limit",
		Title:     "Đạt giới hạn bài/ngày",
		Message:   "Nick " + accountName + " đã đạt giới hạn 20 bài/ngày.",
		AccountID: &accountID,
	}
	return s.CreateNotification(n)
}

// NotifyTokenExpiring tạo thông báo token sắp hết hạn
func (s *Store) NotifyTokenExpiring(accountID string, accountName string, daysLeft int) error {
	n := &Notification{
		Type:      "token_expiring",
		Title:     "Token sắp hết hạn",
		Message:   "Token của nick " + accountName + " sẽ hết hạn trong " + string(rune(daysLeft)) + " ngày. Vui lòng gia hạn.",
		AccountID: &accountID,
	}
	return s.CreateNotification(n)
}

// NotifyPostFailed tạo thông báo đăng bài thất bại
func (s *Store) NotifyPostFailed(accountID string, accountName string, pageName string, reason string) error {
	n := &Notification{
		Type:      "post_failed",
		Title:     "Đăng bài thất bại",
		Message:   "Không thể đăng bài lên " + pageName + " từ nick " + accountName + ": " + reason,
		AccountID: &accountID,
	}
	return s.CreateNotification(n)
}

// NotifyWarningThreshold tạo thông báo đạt 80% giới hạn
func (s *Store) NotifyWarningThreshold(accountID string, accountName string, current, max int) error {
	n := &Notification{
		Type:      "warning_threshold",
		Title:     "Sắp đạt giới hạn bài/ngày",
		Message:   "Nick " + accountName + " đã đăng " + string(rune(current)) + "/" + string(rune(max)) + " bài hôm nay (80%).",
		AccountID: &accountID,
	}
	return s.CreateNotification(n)
}
