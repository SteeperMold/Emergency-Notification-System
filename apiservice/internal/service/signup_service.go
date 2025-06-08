package service

import (
	"context"
	"net/mail"

	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// SignupService provides user registration functionality.
type SignupService struct {
	repository domain.UserRepository
}

// NewSignupService creates and returns a new SignupService with the given user repository.
func NewSignupService(r domain.UserRepository) *SignupService {
	return &SignupService{
		repository: r,
	}
}

// CreateUser validates the credentials, checks for existing users,
// hashes the password, and creates a new user in the repository.
// It returns the created user or an error if any step fails.
func (ss *SignupService) CreateUser(ctx context.Context, userData *domain.SignupRequest) (*models.User, error) {
	_, err := mail.ParseAddress(userData.Email)
	if err != nil {
		return nil, domain.ErrInvalidEmail
	}

	if len(userData.Password) < 5 || len(userData.Password) > 100 {
		return nil, domain.ErrInvalidPassword
	}

	_, err = ss.repository.GetUserByEmail(ctx, userData.Email)
	if err == nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &models.User{
		Email:        userData.Email,
		PasswordHash: string(passwordHash),
	}

	newUser, err = ss.repository.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}
