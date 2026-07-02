package tests

import (
	"testing"

	"github.com/ShreyasDr71/GoPAMS/config"
	"github.com/ShreyasDr71/GoPAMS/services"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretPass123!"
	hash, err := services.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == password {
		t.Fatalf("Hash should not match plaintext password")
	}

	if !services.CheckPasswordHash(password, hash) {
		t.Fatalf("Password verification failed")
	}

	if services.CheckPasswordHash("wrong_password", hash) {
		t.Fatalf("Verification should fail for incorrect password")
	}
}

func TestJWTGenerationAndValidation(t *testing.T) {
	// Initialize minimal config
	config.AppConfig = &config.Config{
		JWTSecret:             "test_secret_key_12345",
		SessionTimeoutMinutes: 5,
	}

	token, err := services.GenerateJWT(42, "testuser", false, true, "Engineer")
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	claims, err := services.ValidateJWT(token)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if claims.UserID != 42 || claims.Username != "testuser" || claims.Role != "Engineer" || !claims.MustChangePassword {
		t.Errorf("Claims mismatch. Got: %+v", claims)
	}
}
