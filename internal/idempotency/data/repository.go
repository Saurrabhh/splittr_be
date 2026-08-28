package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/db/dbgen"
	"github.com/Saurrabhh/splittr_be/internal/idempotency/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type DBRepository struct {
	db *db.DB
	tm *db.TransactionManager
}

// NewRepository creates a new DBRepository instance for idempotency storage.
func NewRepository(database *db.DB, tm *db.TransactionManager) *DBRepository {
	return &DBRepository{
		db: database,
		tm: tm,
	}
}

// Get retrieves an existing idempotency record by user ID and key.
func (r *DBRepository) Get(ctx context.Context, userID, key string) (*domain.IdempotencyRecord, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid: %w", err)
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	row, err := q.GetIdempotencyKey(ctx, dbgen.GetIdempotencyKeyParams{
		UserID: parsedUserID,
		Key:    key,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get idempotency key: %w", err)
	}

	return toDomainRecord(row)
}

// Create stores a new idempotency key. If conflict exists, returns nil record without error.
func (r *DBRepository) Create(ctx context.Context, rec *domain.IdempotencyRecord) (*domain.IdempotencyRecord, error) {
	parsedUserID, err := uuid.Parse(rec.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user uuid: %w", err)
	}

	var headersBytes []byte
	if len(rec.ResponseHeaders) > 0 {
		headersBytes, err = json.Marshal(rec.ResponseHeaders)
		if err != nil {
			return nil, fmt.Errorf("marshal headers: %w", err)
		}
	}

	var respCode pgtype.Int4
	if rec.ResponseCode != nil {
		respCode = pgtype.Int4{Int32: int32(*rec.ResponseCode), Valid: true}
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	row, err := q.CreateIdempotencyKey(ctx, dbgen.CreateIdempotencyKeyParams{
		Key:             rec.Key,
		UserID:          parsedUserID,
		RequestPath:     rec.RequestPath,
		RequestMethod:   rec.RequestMethod,
		RequestHash:     rec.RequestHash,
		ResponseCode:    respCode,
		ResponseHeaders: headersBytes,
		ResponseBody:    rec.ResponseBody,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Conflict: row already exists
			return nil, nil
		}
		return nil, fmt.Errorf("create idempotency key: %w", err)
	}

	return toDomainRecord(row)
}

// UpdateResponse updates the cached HTTP status code, headers, and body for an idempotency key.
func (r *DBRepository) UpdateResponse(ctx context.Context, userID, key string, code int, headers map[string]string, body []byte) error {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user uuid: %w", err)
	}

	var headersBytes []byte
	if len(headers) > 0 {
		var err error
		headersBytes, err = json.Marshal(headers)
		if err != nil {
			return fmt.Errorf("marshal response headers: %w", err)
		}
	}

	client := r.tm.GetTxOrPool(ctx)
	q := dbgen.New(client)

	return q.UpdateIdempotencyKeyResponse(ctx, dbgen.UpdateIdempotencyKeyResponseParams{
		UserID:          parsedUserID,
		Key:             key,
		ResponseCode:    pgtype.Int4{Int32: int32(code), Valid: true},
		ResponseHeaders: headersBytes,
		ResponseBody:    body,
	})
}

func toDomainRecord(row dbgen.IdempotencyKey) (*domain.IdempotencyRecord, error) {
	var headers map[string]string
	if len(row.ResponseHeaders) > 0 {
		if err := json.Unmarshal(row.ResponseHeaders, &headers); err != nil {
			return nil, fmt.Errorf("unmarshal response headers: %w", err)
		}
	}

	var respCode *int
	if row.ResponseCode.Valid {
		c := int(row.ResponseCode.Int32)
		respCode = &c
	}

	return &domain.IdempotencyRecord{
		ID:              row.ID.String(),
		Key:             row.Key,
		UserID:          row.UserID.String(),
		RequestPath:     row.RequestPath,
		RequestMethod:   row.RequestMethod,
		RequestHash:     row.RequestHash,
		ResponseCode:    respCode,
		ResponseHeaders: headers,
		ResponseBody:    row.ResponseBody,
		CreatedAt:       row.CreatedAt.Time,
		LockedAt:        row.LockedAt.Time,
	}, nil
}
