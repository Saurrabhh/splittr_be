package auth

import (
	"context"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
)

// FirebaseVerifier wraps the Firebase Auth Client to implement TokenVerifier.
type FirebaseVerifier struct {
	client *firebaseAuth.Client
}

// NewFirebaseVerifier initializes a Firebase Admin Auth client.
func NewFirebaseVerifier(ctx context.Context, projectID string) (*FirebaseVerifier, error) {
	cfg := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, cfg)
	if err != nil {
		return nil, err
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	return &FirebaseVerifier{client: client}, nil
}

// VerifyIDToken verifies the Firebase ID token.
func (fv *FirebaseVerifier) VerifyIDToken(ctx context.Context, idToken string) (*firebaseAuth.Token, error) {
	return fv.client.VerifyIDToken(ctx, idToken)
}
