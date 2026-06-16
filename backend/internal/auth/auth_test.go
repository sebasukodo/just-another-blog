package auth

import (
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestService(t *testing.T, secret, issuer string) *Service {
	t.Helper()
	return &Service{
		TokenSecret: secret,
		TokenIssuer: issuer,
		Logger:      slog.Default(),
	}
}

func TestAuth(t *testing.T) {
	type test struct {
		name           string
		userID         uuid.UUID
		issuer         string
		secret         string
		validateSecret string
		expiresIn      time.Duration
		tamper         bool
		expectError    bool
	}

	testcases := []test{
		{
			name:           "valid token",
			userID:         uuid.New(),
			issuer:         "testing",
			secret:         "secret",
			validateSecret: "secret",
			expiresIn:      time.Hour,
			expectError:    false,
		},
		{
			name:           "invalid secret",
			userID:         uuid.New(),
			issuer:         "testing",
			secret:         "secret",
			validateSecret: "wrong-secret",
			expiresIn:      time.Hour,
			expectError:    true,
		},
		{
			name:           "expired token",
			userID:         uuid.New(),
			issuer:         "testing",
			secret:         "secret",
			validateSecret: "secret",
			expiresIn:      -1 * time.Minute,
			expectError:    true,
		},
		{
			name:           "tampered token",
			userID:         uuid.New(),
			issuer:         "testing",
			secret:         "secret",
			validateSecret: "secret",
			expiresIn:      time.Hour,
			tamper:         true,
			expectError:    true,
		},
	}

	passCount := 0
	failCount := 0

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			makeSvc := newTestService(t, test.secret, test.issuer)
			token, err := makeSvc.MakeJWT(test.userID, test.expiresIn)
			if err != nil {
				failCount++
				t.Fatalf("MakeJWT failed: %v", err)
			}

			if test.tamper {
				token = token[:len(token)-4] + "xxxx"
			}

			validateSvc := newTestService(t, test.validateSecret, test.issuer)
			userID, err := validateSvc.ValidateJWT(token)

			if test.expectError {
				if err == nil {
					failCount++
					t.Errorf("expected error, got nil")
				} else {
					passCount++
				}
				return
			}

			if err != nil {
				failCount++
				t.Fatalf("ValidateJWT failed: %v", err)
				return
			}

			if userID != test.userID {
				failCount++
				t.Errorf("got userID %v, want %v", userID, test.userID)
				return
			}

			passCount++
		})
	}

	t.Logf("Results: %d passed, %d failed", passCount, failCount)
}
