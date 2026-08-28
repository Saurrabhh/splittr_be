package http

type UpdateNotificationRequest struct {
	IsRead bool `json:"isRead"`
} // @name Notification.UpdateRequest

type BulkUpdateNotificationRequest struct {
	IsRead bool `json:"isRead"`
} // @name Notification.BulkUpdateRequest
