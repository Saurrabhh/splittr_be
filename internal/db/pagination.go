package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ParseCursor converts optional time and string ID pointers into database pagination types.
func ParseCursor(lastTime *time.Time, lastID *string) (pgtype.Timestamptz, uuid.UUID) {
	var pgLastTime pgtype.Timestamptz
	if lastTime != nil {
		pgLastTime = pgtype.Timestamptz{Time: *lastTime, Valid: true}
	}

	var lastIDUUID uuid.UUID
	if lastID != nil {
		if parsed, err := uuid.Parse(*lastID); err == nil {
			lastIDUUID = parsed
		}
	}

	return pgLastTime, lastIDUUID
}
