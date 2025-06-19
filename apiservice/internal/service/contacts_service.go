package service

import (
	"context"
	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/internal/phoneutils"
)

type ContactsService struct {
	repository domain.ContactsRepository
}

func NewContactsService(r domain.ContactsRepository) *ContactsService {
	return &ContactsService{
		repository: r,
	}
}

func (cs *ContactsService) GetContactsByUserID(ctx context.Context, userID int) ([]*models.Contact, error) {
	return cs.repository.GetContactsByUserID(ctx, userID)
}

func (cs *ContactsService) GetContactByID(ctx context.Context, userID int, contactID int) (*models.Contact, error) {
	return cs.repository.GetContactByID(ctx, userID, contactID)
}

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

func (cs *ContactsService) UpdateContact(ctx context.Context, userID int, contactID int, updatedContact *models.Contact) (*models.Contact, error) {
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

func (cs *ContactsService) DeleteContact(ctx context.Context, userID int, contactID int) error {
	return cs.repository.DeleteContact(ctx, userID, contactID)
}
