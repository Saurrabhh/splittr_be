package pagination_test

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

func TestParseCursor_ValidInput(t *testing.T) {
	expected, _ := time.Parse(time.RFC3339, "2026-07-18T18:00:00Z")
	encoded := pagination.EncodeCursor(expected, "abc-123")
	c := pagination.ParseCursor(encoded)
	if c.LastTime == nil || c.LastID == nil {
		t.Fatal("expected non-nil cursor fields")
	}
	if !c.LastTime.Equal(expected) {
		t.Errorf("time mismatch: got %v", *c.LastTime)
	}
	if *c.LastID != "abc-123" {
		t.Errorf("id mismatch: got %s", *c.LastID)
	}
}


func TestParseCursor_EmptyString(t *testing.T) {
	c := pagination.ParseCursor("")
	if c.LastTime != nil || c.LastID != nil {
		t.Fatal("expected zero-value cursor for empty input")
	}
}

func TestParseCursor_MalformedInput(t *testing.T) {
	malformedInputs := []string{
		"notadate_uuid",
		"_uuid",
		"2026-07-18T18:00:00Z",
		"_",
		"2026-07-18T18:00:00Zuuid", // missing underscore
		"2026-99-99T99:99:99Z_123", // bad RFC3339 date
		"2026-07-18T18:00:00Z_",    // missing ID after underscore
		"invalidcursor",
	}

	for _, bad := range malformedInputs {
		c := pagination.ParseCursor(bad)
		if c.LastTime != nil || c.LastID != nil {
			t.Errorf("expected zero cursor for %q, got LastTime=%v, LastID=%v", bad, c.LastTime, c.LastID)
		}
	}
}

func TestEncodeCursor_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	encoded := pagination.EncodeCursor(now, "my-id")
	c := pagination.ParseCursor(encoded)
	if c.LastTime == nil || !c.LastTime.Equal(now) {
		t.Errorf("round-trip time mismatch")
	}
	if c.LastID == nil || *c.LastID != "my-id" {
		t.Errorf("round-trip id mismatch")
	}
}

func TestEncodeCursor_SubSecondPrecision(t *testing.T) {
	now := time.Date(2026, 7, 18, 18, 0, 0, 500000000, time.UTC)
	encoded := pagination.EncodeCursor(now, "my-id")
	c := pagination.ParseCursor(encoded)
	if c.LastTime == nil || !c.LastTime.Equal(now) {
		t.Errorf("expected sub-second precision preserved, got %v", c.LastTime)
	}
	if c.LastID == nil || *c.LastID != "my-id" {
		t.Errorf("round-trip id mismatch")
	}
}

func TestParseParams_DefaultsAndClamping(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=500", nil)
	p := pagination.ParseParams(req, 20, 100)
	if p.Limit != 100 {
		t.Errorf("expected clamped limit 100, got %d", p.Limit)
	}
}

func TestParseParams_ExcessiveLimitClamping(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=9999", nil)
	p := pagination.ParseParams(req, 20, 100)
	if p.Limit != 100 {
		t.Errorf("expected excessive limit to be clamped to maxLimit 100, got %d", p.Limit)
	}
}

func TestParseParams_NegativeLimitFallback(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=-10", nil)
	p := pagination.ParseParams(req, 20, 100)
	if p.Limit != 20 {
		t.Errorf("expected negative limit fallback to defaultLimit 20, got %d", p.Limit)
	}
}

func TestParseParams_DefaultWhenMissing(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	p := pagination.ParseParams(req, 20, 100)
	if p.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", p.Limit)
	}
}

func TestParseParams_ExtractsCursor(t *testing.T) {
	req := httptest.NewRequest("GET", "/?cursor=abc-123", nil)
	p := pagination.ParseParams(req, 20, 100)
	if p.Cursor != "abc-123" {
		t.Errorf("expected cursor abc-123, got %q", p.Cursor)
	}
}

func TestParseParams_ClampsDefaultAboveMax(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	p := pagination.ParseParams(req, 200, 100)
	if p.Limit != 100 {
		t.Errorf("expected default clamped to maxLimit 100, got %d", p.Limit)
	}
}

func TestParseParams_ClampsDefaultBelowMin(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	p := pagination.ParseParams(req, 0, 100)
	if p.Limit != 1 {
		t.Errorf("expected default clamped to min 1, got %d", p.Limit)
	}
}

func TestParseParams_OverflowLimitFallback(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=9999999999999999999", nil)
	p := pagination.ParseParams(req, 20, 100)
	if p.Limit != 20 {
		t.Errorf("expected overflowing limit to fall back to default 20, got %d", p.Limit)
	}
}

func TestBuildResponse_HasMoreTrimming(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6} // 6 items fetched for limit=5
	resp := pagination.BuildResponse(items, 5, func(i int) string { return strconv.Itoa(i) })
	if len(resp.Data) != 5 {
		t.Errorf("expected 5 items, got %d", len(resp.Data))
	}
	if !resp.Pagination.HasMore {
		t.Error("expected HasMore=true")
	}
	if resp.Pagination.NextCursor != "5" {
		t.Errorf("expected NextCursor=5, got %s", resp.Pagination.NextCursor)
	}
}

func TestBuildResponse_NoMore(t *testing.T) {
	items := []int{1, 2}
	resp := pagination.BuildResponse(items, 5, func(i int) string { return strconv.Itoa(i) })
	if resp.Pagination.HasMore {
		t.Error("expected HasMore=false")
	}
	if resp.Pagination.NextCursor != "" {
		t.Errorf("expected empty NextCursor, got %s", resp.Pagination.NextCursor)
	}
}

func TestBuildResponse_ZeroLimit(t *testing.T) {
	items := []int{1, 2, 3}
	resp := pagination.BuildResponse(items, 0, func(i int) string { return strconv.Itoa(i) })
	if len(resp.Data) != 3 {
		t.Errorf("expected all items returned, got %d", len(resp.Data))
	}
	if resp.Pagination.HasMore {
		t.Error("expected HasMore=false for zero limit")
	}
	if resp.Pagination.NextCursor != "" {
		t.Errorf("expected empty NextCursor, got %s", resp.Pagination.NextCursor)
	}
}

func TestBuildResponse_NegativeLimit(t *testing.T) {
	items := []int{1, 2, 3}
	resp := pagination.BuildResponse(items, -1, func(i int) string { return strconv.Itoa(i) })
	if len(resp.Data) != 3 {
		t.Errorf("expected all items returned, got %d", len(resp.Data))
	}
	if resp.Pagination.HasMore {
		t.Error("expected HasMore=false for negative limit")
	}
}

func TestBuildResponse_EmptyItemsNonNullData(t *testing.T) {
	var items []int // nil slice
	resp := pagination.BuildResponse(items, 5, func(i int) string { return strconv.Itoa(i) })
	if resp.Data == nil {
		t.Error("expected Data to be non-nil empty slice for JSON []")
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Data))
	}
	if resp.Pagination.HasMore {
		t.Error("expected HasMore=false for empty items")
	}
	if resp.Pagination.NextCursor != "" {
		t.Errorf("expected empty NextCursor, got %s", resp.Pagination.NextCursor)
	}
}
