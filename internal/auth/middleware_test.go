package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTokenVerifier struct {
	mock.Mock
}

func (m *mockTokenVerifier) VerifyIDToken(ctx context.Context, idToken string) (*firebaseAuth.Token, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*firebaseAuth.Token), args.Error(1)
}

func TestMiddleware_Authenticate(t *testing.T) {
	tests := []struct {
		name             string
		authHeader       string
		setupVerifier    func(*mockTokenVerifier)
		expectedStatus   int
		expectedIdentity *auth.Identity
	}{
		{
			name:       "Valid Bearer token with full claims (UID, email, phone)",
			authHeader: "Bearer valid-full-token",
			setupVerifier: func(mv *mockTokenVerifier) {
				mv.On("VerifyIDToken", mock.Anything, "valid-full-token").Return(&firebaseAuth.Token{
					UID: "test-user-123",
					Claims: map[string]interface{}{
						"email":        "test@example.com",
						"phone_number": "+1234567890",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedIdentity: &auth.Identity{
				UserID: "test-user-123",
				Email:  "test@example.com",
				Phone:  "+1234567890",
			},
		},
		{
			name:             "Missing Authorization header",
			authHeader:       "",
			setupVerifier:    func(mv *mockTokenVerifier) {},
			expectedStatus:   http.StatusUnauthorized,
			expectedIdentity: nil,
		},
		{
			name:             "Malformed Authorization header (missing Bearer prefix)",
			authHeader:       "InvalidTokenWithoutBearer",
			setupVerifier:    func(mv *mockTokenVerifier) {},
			expectedStatus:   http.StatusUnauthorized,
			expectedIdentity: nil,
		},
		{
			name:             "Malformed Authorization header (non-Bearer scheme)",
			authHeader:       "Basic dXNlcjpwYXNz",
			setupVerifier:    func(mv *mockTokenVerifier) {},
			expectedStatus:   http.StatusUnauthorized,
			expectedIdentity: nil,
		},
		{
			name:       "Invalid/Expired token returned by verifier",
			authHeader: "Bearer expired-or-invalid-token",
			setupVerifier: func(mv *mockTokenVerifier) {
				mv.On("VerifyIDToken", mock.Anything, "expired-or-invalid-token").Return(nil, errors.New("token expired or invalid"))
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedIdentity: nil,
		},
		{
			name:       "Valid token missing optional email/phone claims",
			authHeader: "Bearer valid-minimal-token",
			setupVerifier: func(mv *mockTokenVerifier) {
				mv.On("VerifyIDToken", mock.Anything, "valid-minimal-token").Return(&firebaseAuth.Token{
					UID:    "test-user-456",
					Claims: map[string]interface{}{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedIdentity: &auth.Identity{
				UserID: "test-user-456",
				Email:  "",
				Phone:  "",
			},
		},
		{
			name:       "Valid token with sign_in_provider and email_verified",
			authHeader: "Bearer verified-token",
			setupVerifier: func(mv *mockTokenVerifier) {
				mv.On("VerifyIDToken", mock.Anything, "verified-token").Return(&firebaseAuth.Token{
					UID: "test-user-789",
					Firebase: firebaseAuth.FirebaseInfo{
						SignInProvider: "password",
					},
					Claims: map[string]interface{}{
						"email":          "test@example.com",
						"email_verified": true,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedIdentity: &auth.Identity{
				UserID:         "test-user-789",
				Email:          "test@example.com",
				Phone:          "",
				EmailVerified:  true,
				SignInProvider: "password",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mv := new(mockTokenVerifier)
			tt.setupVerifier(mv)

			mw := auth.NewMiddleware(mv)
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				identity := auth.IdentityFrom(r.Context())
				if tt.expectedIdentity != nil {
					assert.NotNil(t, identity)
					assert.Equal(t, tt.expectedIdentity.UserID, identity.UserID)
					assert.Equal(t, tt.expectedIdentity.Email, identity.Email)
					assert.Equal(t, tt.expectedIdentity.Phone, identity.Phone)
					assert.Equal(t, tt.expectedIdentity.EmailVerified, identity.EmailVerified)
					assert.Equal(t, tt.expectedIdentity.SignInProvider, identity.SignInProvider)
				} else {
					assert.Nil(t, identity)
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			mw.Authenticate(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, nextCalled, "next handler should be called on 200 OK")
			} else {
				assert.False(t, nextCalled, "next handler should not be called on error")
			}
			mv.AssertExpectations(t)
		})
	}
}

func TestMiddleware_RequireVerifiedEmail(t *testing.T) {
	tests := []struct {
		name           string
		identity       *auth.Identity
		expectedStatus int
	}{
		{
			name: "Password provider with email_verified=false is rejected",
			identity: &auth.Identity{
				UserID:         "user-1",
				Email:          "test@example.com",
				SignInProvider: "password",
				EmailVerified:  false,
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Password provider with email_verified=true is allowed",
			identity: &auth.Identity{
				UserID:         "user-2",
				Email:          "test@example.com",
				SignInProvider: "password",
				EmailVerified:  true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Google provider with email_verified=false is allowed",
			identity: &auth.Identity{
				UserID:         "user-3",
				Email:          "test@gmail.com",
				SignInProvider: "google.com",
				EmailVerified:  false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Apple provider with email_verified=true is allowed",
			identity: &auth.Identity{
				UserID:         "user-4",
				Email:          "test@privaterelay.appleid.com",
				SignInProvider: "apple.com",
				EmailVerified:  true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Nil identity passes through to downstream auth guard",
			identity:       nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := auth.NewMiddleware(nil)
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.identity != nil {
				req = req.WithContext(auth.WithIdentity(req.Context(), tt.identity))
			}
			rr := httptest.NewRecorder()

			mw.RequireVerifiedEmail(next).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, nextCalled)
			} else {
				assert.False(t, nextCalled)
				var body map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &body)
				assert.NoError(t, err)
				assert.Equal(t, "EMAIL_NOT_VERIFIED", body["errorCode"])
				assert.Equal(t, "EMAIL_NOT_VERIFIED", body["code"])
				assert.Equal(t, float64(403), body["statusCode"])
			}
		})
	}
}

