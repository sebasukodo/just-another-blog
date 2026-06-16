package auth

import (
	"testing"
)

func TestPassword(t *testing.T) {
	type test struct {
		name          string
		password      string
		checkPassword string
		wrongHash     string
		expectMatch   bool
		expectError   bool
	}

	testcase := []test{
		{
			name:          "correct password",
			password:      "testingPW",
			checkPassword: "testingPW",
			expectMatch:   true,
			expectError:   false,
		},
		{
			name:          "wrong password",
			password:      "testingPW",
			checkPassword: "wrongPW",
			expectMatch:   false,
			expectError:   false,
		},
		{
			name:          "invalid hash",
			password:      "testingPW",
			checkPassword: "testingPW",
			wrongHash:     "not-a-valid-hash",
			expectMatch:   false,
			expectError:   true,
		},
	}

	passCount := 0
	failCount := 0

	for _, test := range testcase {
		t.Run(test.name, func(t *testing.T) {
			hashedPW, err := HashPassword(test.password)
			if err != nil {
				failCount++
				t.Fatalf("hashing failed: %v", err)
			}

			hash := hashedPW
			if test.wrongHash != "" {
				hash = test.wrongHash
			}

			ok, err := CheckPasswordHash(test.checkPassword, hash)

			if test.expectError {
				if err == nil {
					failCount++
					t.Errorf("expected error, got nil")
					return
				}
				passCount++
				return
			}

			if err != nil {
				failCount++
				t.Fatalf("CheckPasswordHash failed: %v", err)
				return
			}

			if ok != test.expectMatch {
				failCount++
				t.Errorf("got match=%v, want match=%v", ok, test.expectMatch)
				return
			}

			passCount++
		})
	}

	t.Logf("Results: %d passed, %d failed", passCount, failCount)
}
