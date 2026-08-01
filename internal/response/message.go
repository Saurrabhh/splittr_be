package response

// MessageResponse is the canonical typed response for mutation endpoints
// that have no domain entity to return (e.g. add member, mark notification as read).
// Replaces ad-hoc map[string]string{"message": "..."} across all handlers so that
// Flutter clients can deserialize with a single shared MessageResponse class.
type MessageResponse struct {
	Message string `json:"message"`
} // @name MessageResponse
