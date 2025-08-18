package domain

import (
	"context"
	"fmt"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
)

var (
	// ErrUserNotExists is returned when a requested user cannot be found in the repository.
	ErrUserNotExists = fmt.Errorf("user doesn't exist")
)

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
}
