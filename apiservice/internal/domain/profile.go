package domain

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
)

// ProfileService defines the interface for retrieving user profile information.
type ProfileService interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
}
