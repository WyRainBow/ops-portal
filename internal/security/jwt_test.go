package security

import (
	"os"
	"testing"
	"time"
)

func TestCreateAndParseToken(t *testing.T) {
	// Set a test JWT secret
	os.Setenv("OPS_PORTAL_JWT_SECRET", "test-secret-for-unit-tests")
	defer os.Unsetenv("OPS_PORTAL_JWT_SECRET")

	userID := int64(123)
	username := "testuser"
	role := "admin"

	// Create token
	token, err := CreateToken(userID, username, role)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("CreateToken returned empty token")
	}

	// Parse token
	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}

	if claims.Role != role {
		t.Errorf("Expected role %s, got %s", role, claims.Role)
	}

	if claims.Subject != "123" {
		t.Errorf("Expected subject 123, got %s", claims.Subject)
	}
}

func TestParseInvalidToken(t *testing.T) {
	os.Setenv("OPS_PORTAL_JWT_SECRET", "test-secret-for-unit-tests")
	defer os.Unsetenv("OPS_PORTAL_JWT_SECRET")

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid token", "invalid.token.string"},
		{"wrong secret", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3QiLCJyb2xlIjoiYWRtaW4ifQ.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseToken(tt.token)
			if err == nil {
				t.Error("Expected error for invalid token, got nil")
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	os.Setenv("OPS_PORTAL_JWT_SECRET", "test-secret-for-unit-tests")
	os.Setenv("OPS_PORTAL_JWT_EXPIRE_HOURS", "1")
	defer os.Unsetenv("OPS_PORTAL_JWT_SECRET")
	defer os.Unsetenv("OPS_PORTAL_JWT_EXPIRE_HOURS")

	token, err := CreateToken(1, "user", "member")
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	now := time.Now().UTC()
	expectedExpiry := now.Add(1 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	// Allow 1 minute difference for test execution time
	diff := actualExpiry.Sub(expectedExpiry)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("Token expiry time is off by more than 1 minute: expected ~%v, got %v", expectedExpiry, actualExpiry)
	}
}
