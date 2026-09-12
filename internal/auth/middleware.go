package auth

import (
	"context"
	"net/http"
	"strings"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/Saurrabhh/splittr_be/internal/response"
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
			response.Unauthorized(w, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorized(w, "invalid authorization header format")
			return
		}

		idToken := parts[1]
		token, err := m.verifier.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			response.Unauthorized(w, "invalid token")
			return
		}

		identity := parseIdentity(token)
		ctx := WithIdentity(r.Context(), identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthenticate extracts and verifies Bearer token if present, but proceeds even if missing or invalid.
func (m *Middleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				token, err := m.verifier.VerifyIDToken(r.Context(), parts[1])
				if err == nil {
					identity := parseIdentity(token)
					r = r.WithContext(WithIdentity(r.Context(), identity))
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireVerifiedEmail blocks requests from password-authenticated users whose email is not verified.
func (m *Middleware) RequireVerifiedEmail(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := IdentityFrom(r.Context())
		if identity != nil && identity.SignInProvider == "password" && !identity.EmailVerified {
			response.EmailNotVerified(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func parseIdentity(token *firebaseAuth.Token) *Identity {
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

	var emailVerified bool
	if evVal, ok := token.Claims["email_verified"]; ok {
		if evBool, ok := evVal.(bool); ok {
			emailVerified = evBool
		}
	}

	signInProvider := token.Firebase.SignInProvider
	if signInProvider == "" {
		if fbMap, ok := token.Claims["firebase"].(map[string]interface{}); ok {
			if p, ok := fbMap["sign_in_provider"].(string); ok {
				signInProvider = p
			}
		}
	}

	return &Identity{
		UserID:         token.UID,
		Email:          email,
		Phone:          phone,
		EmailVerified:  emailVerified,
		SignInProvider: signInProvider,
	}
}
