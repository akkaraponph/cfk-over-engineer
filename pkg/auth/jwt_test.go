package auth

import (
	"os"
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	token, err := GenerateToken("user-1", "tenant-1", "admin", "test@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("expected user ID 'user-1', got '%s'", claims.UserID)
	}
	if claims.TenantID != "tenant-1" {
		t.Errorf("expected tenant ID 'tenant-1', got '%s'", claims.TenantID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role 'admin', got '%s'", claims.Role)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", claims.Email)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	_, err := ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret-1")
	token, _ := GenerateToken("user-1", "tenant-1", "admin", "test@example.com")

	os.Setenv("JWT_SECRET", "secret-2")
	_, err := ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestGenerateToken_MissingSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	_, err := GenerateToken("user-1", "tenant-1", "admin", "test@example.com")
	if err == nil {
		t.Fatal("expected error when JWT_SECRET not set")
	}
	if err != ErrMissingSecret {
		t.Errorf("expected ErrMissingSecret, got %v", err)
	}
}
