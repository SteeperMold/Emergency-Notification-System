package service

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/phoneutils"
)

// ContactsService encapsulates business logic around contacts management.
// It validates input and delegates persistence to a ContactsRepository.
type ContactsService struct {
	repository   domain.ContactsRepository
	defaultLimit int
	maxLimit     int
}

// NewContactsService constructs a ContactsService given a repository implementation.
func NewContactsService(r domain.ContactsRepository, defaultLimit, maxLimit int) *ContactsService {
	return &ContactsService{
		repository:   r,
		defaultLimit: defaultLimit,
		maxLimit:     maxLimit,
	}
}

// GetContactsByUserID retrieves all contacts for a given user.
func (cs *ContactsService) GetContactsByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Contact, error) {
	if limit <= 0 {
		limit = cs.defaultLimit
	}
	if limit > cs.maxLimit {
		limit = cs.maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return cs.repository.GetContactsByUserID(ctx, userID, limit, offset)
}

// GetContactByID retrieves a single contact by its ID for a given user.
func (cs *ContactsService) GetContactByID(ctx context.Context, userID, contactID int) (*models.Contact, error) {
	return cs.repository.GetContactByID(ctx, userID, contactID)
}

// CreateContact validates the incoming contact, formats its phone number, and then creates it via the repository.
// Returns the created Contact model or a domain error on validation or persistence failure.
func (cs *ContactsService) CreateContact(ctx context.Context, contact *models.Contact) (*models.Contact, error) {
	if len(contact.Name) == 0 || len(contact.Name) > 32 {
		return nil, domain.ErrInvalidContactName
	}

	normalizedNum, err := phoneutils.FormatToE164(contact.Phone, phoneutils.RegionRU)
	if err != nil {
		return nil, domain.ErrInvalidContactPhone
	}

	contact.Phone = normalizedNum

	return cs.repository.CreateContact(ctx, contact)
}

// UpdateContact validates and formats the updated contact, then applies changes via repository.
func (cs *ContactsService) UpdateContact(ctx context.Context, userID, contactID int, updatedContact *models.Contact) (*models.Contact, error) {
	if len(updatedContact.Name) == 0 || len(updatedContact.Name) > 32 {
		return nil, domain.ErrInvalidContactName
	}

	normalizedNum, err := phoneutils.FormatToE164(updatedContact.Phone, phoneutils.RegionRU)
	if err != nil {
		return nil, domain.ErrInvalidContactPhone
	}

	updatedContact.Phone = normalizedNum

	return cs.repository.UpdateContact(ctx, userID, contactID, updatedContact)
}

// DeleteContact removes a contact by ID for the specified user.
// Returns an error if deletion fails or the contact does not exist.
func (cs *ContactsService) DeleteContact(ctx context.Context, userID, contactID int) error {
	return cs.repository.DeleteContact(ctx, userID, contactID)
}
