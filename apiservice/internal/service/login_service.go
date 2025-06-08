package service

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// LoginService provides user authentication related operations.
type LoginService struct {
	repository domain.UserRepository
}

// NewLoginService creates a new LoginService with the given UserRepository.
func NewLoginService(r domain.UserRepository) *LoginService {
	return &LoginService{
		repository: r,
	}
}

// GetUserByEmail retrieves a user by their email address.
func (ls *LoginService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return ls.repository.GetUserByEmail(ctx, email)
}

// CompareCredentials compares the stored hashed password of the user with the provided password from login request.
// Returns true if the password matches, false otherwise.
func (ls *LoginService) CompareCredentials(user *models.User, request *domain.LoginRequest) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	return err == nil
}
