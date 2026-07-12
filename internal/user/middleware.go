package user

import (
	"context"
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/response"
)

type contextKey struct{}

var userCtxKey = contextKey{}

// WithUser injects the User into the context.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// From retrieves the User from the context.
// Returns nil if no User is present.
func From(ctx context.Context) *User {
	u, _ := ctx.Value(userCtxKey).(*User)
	return u
}

// MustFrom retrieves the User from the context, or panics if the User is not present.
// Use this in handlers where the User is guaranteed to exist by the UserContext middleware.
func MustFrom(ctx context.Context) *User {
	u := From(ctx)
	if u == nil {
		panic("user missing from context; ensure UserContext middleware is registered for this route")
	}
	return u
}

// UserContext resolves the Firebase UID from the auth.Identity in the context
// to the local database User and injects it into the request context.
func (h *Handler) UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := auth.IdentityFrom(r.Context())
		if identity == nil {
			response.Unauthorized(w, "unauthorized")
			return
		}

		u, err := h.uc.GetUserByFirebaseUID(r.Context(), identity.UserID)
		if err != nil {
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
