package domain

import (
	"context"
	"fmt"

	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
)

var (
	// ErrEmailAlreadyExists is returned when a user tries to sign up with an email that is already registered.
	ErrEmailAlreadyExists = fmt.Errorf("user with given email already exists")

	// ErrInvalidEmail is returned when the provided email address does not have a valid format.
	ErrInvalidEmail = fmt.Errorf("invalid email")

	// ErrInvalidPassword is returned when the provided password does not meet length requirements.
	ErrInvalidPassword = fmt.Errorf("password is too short or too long")
)

// SignupService defines the behavior for user registration.
type SignupService interface {
	CreateUser(ctx context.Context, userData *SignupRequest) (*models.User, error)
}

// SignupRequest represents the data required to register a new user.
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupResponse represents the response returned after a successful signup.
type SignupResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
}
