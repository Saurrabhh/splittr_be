package pagination

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Cursor holds the decoded position from a cursor string.
type Cursor struct {
	LastTime *time.Time
	LastID   *string
}

// ParseCursor decodes a cursor string of the form "<RFC3339>_<uuid>".
// Returns a zero-value Cursor (both fields nil) on any parse error.
func ParseCursor(s string) Cursor {
	if s == "" {
		return Cursor{}
	}
	// Split on the first underscore that follows the RFC3339 timestamp.
	idx := strings.Index(s, "_")
	if idx < 1 || idx == len(s)-1 {
		return Cursor{}
	}
	t, err := time.Parse(time.RFC3339, s[:idx])
	if err != nil {
		return Cursor{}
	}
	id := s[idx+1:]
	return Cursor{LastTime: &t, LastID: &id}
}

// EncodeCursor serialises a (time, id) pair into a cursor string.
func EncodeCursor(t time.Time, id string) string {
	return t.UTC().Format(time.RFC3339) + "_" + id
}

// Params holds the parsed query parameters for a paginated request.
type Params struct {
	Limit  int32
	Cursor string
}

// ParseParams reads "limit" and "cursor" from the request query string.
// limit is clamped to [1, maxLimit]. Defaults to defaultLimit when absent or invalid.
func ParseParams(r *http.Request, defaultLimit, maxLimit int32) Params {
	limit := defaultLimit
	if s := r.URL.Query().Get("limit"); s != "" {
		if l, err := strconv.Atoi(s); err == nil && l > 0 {
			limit = int32(l)
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return Params{
		Limit:  limit,
		Cursor: r.URL.Query().Get("cursor"),
	}
}

// Meta is the pagination envelope returned in every paginated API response.
type Meta struct {
	NextCursor string `json:"nextCursor"`
	HasMore    bool   `json:"hasMore"`
}

// Response is the standard paginated response wrapper used by all domains.
type Response[T any] struct {
	Data       []T  `json:"data"`
	Pagination Meta `json:"pagination"`
}

// BuildResponse trims the N+1 sentinel item from items (callers query limit+1)
// and builds the paginated response. encodeCursorFn extracts the cursor from
// the last item in the trimmed slice.
func BuildResponse[T any](items []T, requestedLimit int32, encodeCursorFn func(T) string) Response[T] {
	hasMore := int32(len(items)) > requestedLimit
	data := items
	if hasMore {
		data = items[:requestedLimit]
	}
	nextCursor := ""
	if hasMore && len(data) > 0 {
		nextCursor = encodeCursorFn(data[len(data)-1])
	}
	return Response[T]{
		Data: data,
		Pagination: Meta{
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
	}
}
