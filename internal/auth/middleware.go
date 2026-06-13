package auth

import (
	"context"
	"net/http"
	"strings"

	firebaseAuth "firebase.google.com/go/v4/auth"
)

// TokenVerifier defines the interface for verifying auth tokens.
type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*firebaseAuth.Token, error)
}

// Middleware handles authentication verification.
type Middleware struct {
	verifier TokenVerifier
}

// NewMiddleware creates a new Middleware instance.
func NewMiddleware(verifier TokenVerifier) *Middleware {
	return &Middleware{verifier: verifier}
}

// Authenticate extracts the Bearer token, verifies it, and injects the identity into the context.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		idToken := parts[1]
		token, err := m.verifier.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		var email string
		if emailVal, ok := token.Claims["email"]; ok {
			if emailStr, ok := emailVal.(string); ok {
				email = emailStr
			}
		}

		var phone string
		if phoneVal, ok := token.Claims["phone_number"]; ok {
			if phoneStr, ok := phoneVal.(string); ok {
				phone = phoneStr
			}
		}

		identity := &Identity{
			UserID: token.UID,
			Email:  email,
			Phone:  phone,
		}

		ctx := WithIdentity(r.Context(), identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
