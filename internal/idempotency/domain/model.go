package domain

import "time"

// IdempotencyRecord stores the execution lifecycle and cached HTTP response for a unique request key.
type IdempotencyRecord struct {
	ID              string            `json:"id"`
	Key             string            `json:"key"`
	UserID          string            `json:"userId"`
	RequestPath     string            `json:"requestPath"`
	RequestMethod   string            `json:"requestMethod"`
	RequestHash     string            `json:"requestHash"`
	ResponseCode    *int              `json:"responseCode,omitempty"`
	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody    []byte            `json:"responseBody,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	LockedAt        time.Time         `json:"lockedAt"`
}
