package domain

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
)

// LoginService defines the interface for user login operations.
type LoginService interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CompareCredentials(user *models.User, request *LoginRequest) bool
}

// LoginRequest represents the data required to perform a user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse contains the authenticated user details and JWT tokens.
type LoginResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
}
