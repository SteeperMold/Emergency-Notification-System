package domain

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
)

// RefreshTokenService defines the interface for operations related to refresh tokens.
type RefreshTokenService interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
}

// RefreshTokenRequest represents the payload for requesting a new access token using a refresh token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenResponse represents the response containing new access and refresh tokens.
type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
