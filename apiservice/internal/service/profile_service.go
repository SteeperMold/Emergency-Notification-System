package service

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
)

// ProfileService handles retrieval of user profile data.
type ProfileService struct {
	repository domain.UserRepository
}

// NewProfileService returns a new instance of ProfileService using the provided UserRepository.
func NewProfileService(r domain.UserRepository) *ProfileService {
	return &ProfileService{
		repository: r,
	}
}

// GetUserByID fetches a user by their ID from the underlying repository.
// It returns the user model or an error if the user is not found or a retrieval issue occurs.
func (ps *ProfileService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	return ps.repository.GetUserByID(ctx, id)
}
