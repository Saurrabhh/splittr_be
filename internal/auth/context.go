package auth

import (
	"context"
)

// Identity represents the authenticated user's information.
type Identity struct {
	UserID string
	Email  string
	Phone  string
}

// contextKey is the private key type for storing the Identity in the context.
type contextKey struct{}

var identityKey = contextKey{}

// WithIdentity injects the Identity into the context.
func WithIdentity(ctx context.Context, id *Identity) context.Context {
	return context.WithValue(ctx, identityKey, id)
}

// IdentityFrom retrieves the Identity from the context.
// Returns nil if no Identity is present.
func IdentityFrom(ctx context.Context) *Identity {
	id, _ := ctx.Value(identityKey).(*Identity)
	return id
}
