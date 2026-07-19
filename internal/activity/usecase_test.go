package activity

import (
	"context"
	"testing"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/pagination"
)

type mockRepository struct {
	listGroupFeedMock func(groupID string, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Activity, error)
}

func (m *mockRepository) CreateActivity(ctx context.Context, act *Activity) error { return nil }
func (m *mockRepository) CreateActivityVisibility(ctx context.Context, activityID string, userID string) error { return nil }
func (m *mockRepository) ListUserActivities(ctx context.Context, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Activity, error) { return nil, nil }
func (m *mockRepository) ListGroupFeed(ctx context.Context, groupID string, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Activity, error) {
	return m.listGroupFeedMock(groupID, userID, limit, lastTime, lastID)
}

func TestGetGroupFeed_CursorParsing(t *testing.T) {
	mockRepo := &mockRepository{
		listGroupFeedMock: func(groupID string, userID string, limit int32, lastTime *time.Time, lastID *string) ([]Activity, error) {
			if lastTime == nil || lastID == nil {
				t.Fatal("expected non-nil cursor parameters")
			}
			if *lastID != "uuid-test" {
				t.Errorf("expected lastID to be uuid-test, got %s", *lastID)
			}
			expectedTime, _ := time.Parse(time.RFC3339, "2026-07-18T18:00:00Z")
			if !lastTime.Equal(expectedTime) {
				t.Errorf("expected lastTime %v, got %v", expectedTime, *lastTime)
			}
			return []Activity{}, nil
		},
	}

	u := NewUsecase(mockRepo)
	_, err := u.GetGroupFeed(context.Background(), "user-uuid", "group-uuid", pagination.Params{Limit: 10, Cursor: "2026-07-18T18:00:00Z_uuid-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
