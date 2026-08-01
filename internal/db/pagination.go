package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ParseCursor converts optional time and string ID pointers into database pagination types.
// It returns an error when a non-nil lastID is not a valid UUID.
func ParseCursor(lastTime *time.Time, lastID *string) (pgtype.Timestamptz, uuid.UUID, error) {
	var pgLastTime pgtype.Timestamptz
	if lastTime != nil {
		pgLastTime = pgtype.Timestamptz{Time: *lastTime, Valid: true}
	}

	var lastIDUUID uuid.UUID
	if lastID != nil {
		parsed, err := uuid.Parse(*lastID)
		if err != nil {
			return pgtype.Timestamptz{}, uuid.UUID{}, fmt.Errorf("invalid cursor id: %w", err)
		}
		lastIDUUID = parsed
	}

	return pgLastTime, lastIDUUID, nil
}
