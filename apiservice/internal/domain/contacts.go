package domain

import (
	"context"
	"fmt"
	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
)

var (
	ErrContactNotExists = fmt.Errorf("contact doesn't exist")

	ErrInvalidContact      = fmt.Errorf("invalid contact")
	ErrInvalidContactName  = fmt.Errorf("%w: invalid name", ErrInvalidContact)
	ErrInvalidContactPhone = fmt.Errorf("%w: invalid phone", ErrInvalidContact)

	ErrContactAlreadyExists = fmt.Errorf("contact already exists")
)

type ContactsRepository interface {
	GetContactsByUserID(ctx context.Context, userID int) ([]*models.Contact, error)
	GetContactByID(ctx context.Context, userID, contactID int) (*models.Contact, error)
	CreateContact(ctx context.Context, contact *models.Contact) (*models.Contact, error)
	UpdateContact(ctx context.Context, userID, contactID int, updatedContact *models.Contact) (*models.Contact, error)
	DeleteContact(ctx context.Context, userID, contactID int) error
}

type ContactsService interface {
	GetContactsByUserID(ctx context.Context, userID int) ([]*models.Contact, error)
	GetContactByID(ctx context.Context, userID, contactID int) (*models.Contact, error)
	CreateContact(ctx context.Context, contact *models.Contact) (*models.Contact, error)
	UpdateContact(ctx context.Context, userID, contactID int, updatedContact *models.Contact) (*models.Contact, error)
	DeleteContact(ctx context.Context, userID, contactID int) error
}

type PostContactRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type PutContactRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}
