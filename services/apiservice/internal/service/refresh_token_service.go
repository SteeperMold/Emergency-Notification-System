package service

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
)

// RefreshTokenService provides functionality to retrieve user data using their ID,
// typically as part of handling refresh token operations.
type RefreshTokenService struct {
	repository domain.UserRepository
}

// NewRefreshTokenService creates a new instance of RefreshTokenService using the given UserRepository.
func NewRefreshTokenService(r domain.UserRepository) *RefreshTokenService {
	return &RefreshTokenService{
		repository: r,
	}
}

// GetUserByID retrieves a user by their ID from the repository.
// It returns the user model or an error if the user does not exist or the query fails.
func (rts *RefreshTokenService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	return rts.repository.GetUserByID(ctx, id)
}
