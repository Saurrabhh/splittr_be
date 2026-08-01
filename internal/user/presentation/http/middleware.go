package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user/domain"
)

type contextKey struct{}

var userCtxKey = contextKey{}

// WithUser injects the User into the context.
func WithUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// From retrieves the User from the context.
func From(ctx context.Context) *domain.User {
	u, _ := ctx.Value(userCtxKey).(*domain.User)
	return u
}

// MustFrom retrieves the User from the context, or panics if the User is not present.
func MustFrom(ctx context.Context) *domain.User {
	u := From(ctx)
	if u == nil {
		panic("user missing from context; ensure UserContext middleware is registered for this route")
	}
	return u
}

// UserContext resolves the Firebase UID from the auth.Identity in the context to the local database User.
func (h *Handler) UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := auth.IdentityFrom(r.Context())
		if identity == nil {
			response.Unauthorized(w, "unauthorized")
			return
		}

		u, err := h.uc.GetUserByFirebaseUID(r.Context(), identity.UserID)
		if err != nil {
			var appErr *response.AppError
			if errors.As(err, &appErr) && appErr.Type == response.TypeNotFound {
				response.Error(w, http.StatusForbidden, response.ErrUserNotFound, "user registration required")
				return
			}
			response.InternalServerError(w, "failed to resolve local user: "+err.Error())
			return
		}
		if u == nil {
			response.Error(w, http.StatusForbidden, response.ErrUserNotFound, "user registration required")
			return
		}

		ctx := WithUser(r.Context(), u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
