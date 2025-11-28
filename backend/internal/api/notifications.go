package api

import (
	"net/http"

	"fbscheduler/internal/db"

	"github.com/gorilla/mux"
)

// ============================================
// NOTIFICATIONS API
// ============================================

// GetNotifications GET /api/notifications - Danh sách thông báo
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	limit := getQueryInt(r, "limit", 50)
	offset := getQueryInt(r, "offset", 0)
	unreadOnly := r.URL.Query().Get("unread") == "true"

	var notifications []db.Notification
	var err error

	if unreadOnly {
		notifications, err = h.store.GetUnreadNotifications()
	} else {
		notifications, err = h.store.GetAllNotifications(limit, offset)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch notifications: "+err.Error())
		return
	}

	// Get unread count
	unreadCount, _ := h.store.GetUnreadNotificationCount()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"unread_count":  unreadCount,
	})
}

// GetUnreadCount GET /api/notifications/count - Số thông báo chưa đọc
func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.GetUnreadNotificationCount()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get count: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"count": count})
}

// MarkNotificationRead PUT /api/notifications/:id/read - Đánh dấu đã đọc
func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.store.MarkNotificationAsRead(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to mark as read: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Marked as read"})
}

// MarkAllNotificationsRead PUT /api/notifications/read-all - Đánh dấu tất cả đã đọc
func (h *Handler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if err := h.store.MarkAllNotificationsAsRead(); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to mark all as read: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "All marked as read"})
}

// DeleteNotification DELETE /api/notifications/:id - Xóa thông báo
func (h *Handler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.store.DeleteNotification(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete notification: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Notification deleted"})
}
