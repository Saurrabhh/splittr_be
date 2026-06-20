package auth

import (
	"context"
	"os"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// FirebaseVerifier wraps the Firebase Auth Client to implement TokenVerifier.
type FirebaseVerifier struct {
	client *firebaseAuth.Client
}

// NewFirebaseVerifier initializes a Firebase Admin Auth client.
func NewFirebaseVerifier(ctx context.Context) (*FirebaseVerifier, error) {
	var opts []option.ClientOption

	// If the service account JSON is injected directly in the env (e.g. on Vercel)
	if saJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON"); saJSON != "" {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(saJSON)))
	}

	app, err := firebase.NewApp(ctx, nil, opts...)
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
