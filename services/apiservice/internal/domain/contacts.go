package domain

import (
	"context"
	"fmt"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
)

var (
	// ErrContactNotExists indicates a lookup or deletion on a non-existent contact.
	ErrContactNotExists = fmt.Errorf("contact doesn't exist")
	// ErrInvalidContact is the base error for contact validation failures.
	ErrInvalidContact = fmt.Errorf("invalid contact")
	// ErrInvalidContactName indicates the contact's name is empty or too long.
	ErrInvalidContactName = fmt.Errorf("%w: invalid name", ErrInvalidContact)
	// ErrInvalidContactPhone indicates the contact's phone number failed validation.
	ErrInvalidContactPhone = fmt.Errorf("%w: invalid phone", ErrInvalidContact)
	// ErrContactAlreadyExists indicates a uniqueness constraint violation on create/update.
	ErrContactAlreadyExists = fmt.Errorf("contact already exists")
)

// ContactsRepository defines CRUD operations against the persistence layer.
// Implementations should handle SQL details and map domain errors.
type ContactsRepository interface {
	GetAllContactsByUserID(ctx context.Context, userID int) ([]*models.Contact, error)
	GetContactsCountByUserID(ctx context.Context, userID int) (int, error)
	GetContactsPageByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Contact, error)
	GetContactByID(ctx context.Context, userID, contactID int) (*models.Contact, error)
	CreateContact(ctx context.Context, contact *models.Contact) (*models.Contact, error)
	UpdateContact(ctx context.Context, userID, contactID int, updatedContact *models.Contact) (*models.Contact, error)
	DeleteContact(ctx context.Context, userID, contactID int) error
}

// ContactsService defines business logic methods for contacts.
// It validates input and delegates persistence to ContactsRepository.
type ContactsService interface {
	GetContactsCountByUserID(ctx context.Context, userID int) (int, error)
	GetContactsPageByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Contact, error)
	GetContactByID(ctx context.Context, userID, contactID int) (*models.Contact, error)
	CreateContact(ctx context.Context, contact *models.Contact) (*models.Contact, error)
	UpdateContact(ctx context.Context, userID, contactID int, updatedContact *models.Contact) (*models.Contact, error)
	DeleteContact(ctx context.Context, userID, contactID int) error
}

// PostContactRequest defines the payload for creating a new contact via API.
type PostContactRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// PutContactRequest defines the payload for updating an existing contact.
type PutContactRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// GetContactsResponse represents the response payload for getting the list of user's contacts.
type GetContactsResponse struct {
	Contacts []*models.Contact `json:"contacts"`
	Total    int               `json:"total"`
}
