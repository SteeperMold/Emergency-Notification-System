package tokenutils

import (
	"fmt"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

// CreateAccessToken generates a signed JWT access token for a given user using the given secret.
func CreateAccessToken(user *models.User, secret string, expiry time.Duration) (string, error) {
	now := time.Now()

	accessClaims := jwt.MapClaims{
		"email": user.Email,
		"id":    user.ID,
		"nbf":   now.Unix(),
		"exp":   now.Add(expiry).Unix(),
		"iat":   now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return t, nil
}

// CreateRefreshToken generates a signed JWT refresh token for a given user using the given secret.
func CreateRefreshToken(user *models.User, secret string, expiry time.Duration) (string, error) {
	now := time.Now()

	refreshClaims := jwt.MapClaims{
		"id":  user.ID,
		"nbf": now.Unix(),
		"exp": now.Add(expiry).Unix(),
		"iat": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return t, nil
}

// IsAuthorized verifies the validity of a JWT using the given secret.
// It returns true if the token is valid and signed correctly, false otherwise.
func IsAuthorized(authToken string, secret string) (bool, error) {
	_, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

// ExtractIDFromToken extracts the user ID from JWT token using the given secret.
func ExtractIDFromToken(authToken string, secret string) (int, error) {
	token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	idFloat, ok := claims["id"].(float64)
	if !ok {
		return 0, fmt.Errorf("id claim not found or invalid")
	}

	return int(idFloat), nil
}
