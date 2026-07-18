package pagination_test

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

func TestParseCursor_ValidInput(t *testing.T) {
	c := pagination.ParseCursor("2026-07-18T18:00:00Z_abc-123")
	if c.LastTime == nil || c.LastID == nil {
		t.Fatal("expected non-nil cursor fields")
	}
	expected, _ := time.Parse(time.RFC3339, "2026-07-18T18:00:00Z")
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
	for _, bad := range []string{"notadate_uuid", "_uuid", "2026-07-18T18:00:00Z", "_"} {
		c := pagination.ParseCursor(bad)
		if c.LastTime != nil && c.LastID != nil {
			t.Errorf("expected zero cursor for %q", bad)
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

func TestParseParams_DefaultsAndClamping(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=500", nil)
	p := pagination.ParseParams(req, 20, 100)
	if p.Limit != 100 {
		t.Errorf("expected clamped limit 100, got %d", p.Limit)
	}
}

func TestParseParams_DefaultWhenMissing(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	p := pagination.ParseParams(req, 20, 100)
	if p.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", p.Limit)
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
