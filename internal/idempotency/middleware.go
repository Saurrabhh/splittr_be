package idempotency

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/idempotency/domain"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
)

type Middleware struct {
	repo domain.Repository
}

func NewMiddleware(repo domain.Repository) *Middleware {
	return &Middleware{repo: repo}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Handler returns an HTTP middleware that handles deduplication via X-Idempotency-Key.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idempotencyKey := r.Header.Get("X-Idempotency-Key")
		if idempotencyKey == "" {
			idempotencyKey = r.Header.Get("Idempotency-Key")
		}

		// Passthrough if no idempotency key was supplied
		if idempotencyKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		if len(idempotencyKey) > 255 {
			response.BadRequest(w, "Idempotency key length must not exceed 255 characters")
			return
		}

		currUser := user.From(r.Context())
		if currUser == nil {
			next.ServeHTTP(w, r)
			return
		}

		var bodyBytes []byte
		if r.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(r.Body)
			if err != nil {
				response.BadRequest(w, "Failed to read request body")
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		hasher := sha256.New()
		hasher.Write([]byte(r.Method + " " + r.URL.Path + "\n"))
		hasher.Write(bodyBytes)
		reqHash := hex.EncodeToString(hasher.Sum(nil))

		existing, err := m.repo.Get(r.Context(), currUser.ID, idempotencyKey)
		if err != nil {
			response.InternalServerError(w, "Failed to verify idempotency key")
			return
		}

		if existing != nil {
			if existing.RequestHash != reqHash {
				response.Error(w, http.StatusUnprocessableEntity, "IDEMPOTENCY_CONFLICT", "Idempotency key reuse with mismatched request payload")
				return
			}

			if existing.ResponseCode != nil {
				for k, v := range existing.ResponseHeaders {
					w.Header().Set(k, v)
				}
				w.Header().Set("X-Idempotency-Hit", "true")
				w.WriteHeader(*existing.ResponseCode)
				_, _ = w.Write(existing.ResponseBody)
				return
			}

			if time.Since(existing.LockedAt) < 30*time.Second {
				response.Error(w, http.StatusConflict, "CONCURRENT_REQUEST", "A request with this idempotency key is currently being processed")
				return
			}
		}

		rec := &domain.IdempotencyRecord{
			Key:           idempotencyKey,
			UserID:        currUser.ID,
			RequestPath:   r.URL.Path,
			RequestMethod: r.Method,
			RequestHash:   reqHash,
		}
		_, err = m.repo.Create(r.Context(), rec)
		if err != nil {
			response.InternalServerError(w, "Failed to lock idempotency key")
			return
		}

		recorder := newResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		headersToSave := make(map[string]string)
		for k, v := range recorder.Header() {
			if strings.EqualFold(k, "Content-Type") || strings.EqualFold(k, "ETag") {
				if len(v) > 0 {
					headersToSave[k] = v[0]
				}
			}
		}

		_ = m.repo.UpdateResponse(r.Context(), currUser.ID, idempotencyKey, recorder.statusCode, headersToSave, recorder.body.Bytes())
	})
}
