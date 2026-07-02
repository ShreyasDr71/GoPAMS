package services

import (
	"errors"
	"time"

	"github.com/ShreyasDr71/GoPAMS/config"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims defines the structure of JWT custom claims
type JWTClaims struct {
	UserID             uint   `json:"user_id"`
	Username           string `json:"username"`
	IsAdmin            bool   `json:"is_admin"`
	MustChangePassword bool   `json:"must_change_password"`
	Role               string `json:"role"`
	jwt.RegisteredClaims
}

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a hashed password with its plain text version
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates a JWT token for a given user details
func GenerateJWT(userID uint, username string, isAdmin bool, mustChangePassword bool, roleName string) (string, error) {
	jwtSecret := []byte(config.AppConfig.JWTSecret)
	expirationTime := time.Now().Add(time.Duration(config.AppConfig.SessionTimeoutMinutes) * time.Minute)

	claims := &JWTClaims{
		UserID:             userID,
		Username:           username,
		IsAdmin:            isAdmin,
		MustChangePassword: mustChangePassword,
		Role:               roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT parses and validates a JWT token string
func ValidateJWT(tokenStr string) (*JWTClaims, error) {
	jwtSecret := []byte(config.AppConfig.JWTSecret)

	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Enforce expiry check
		if claims.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("token has expired")
		}
		return claims, nil
	}

	return nil, errors.New("invalid claims")
}
