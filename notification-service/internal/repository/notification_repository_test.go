//go:build integration
// +build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNotificationRepository_CreateMultipleNotifications(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewNotificationRepository(testPool)

	ntf1 := &models.Notification{
		ID:             uuid.New(),
		UserID:         101,
		Text:           "First test",
		RecipientPhone: "+10000000001",
	}
	ntf2 := &models.Notification{
		ID:             uuid.New(),
		UserID:         101,
		Text:           "Second test",
		RecipientPhone: "+10000000002",
	}

	err := repo.CreateMultipleNotifications(ctx, []*models.Notification{ntf1, ntf2})
	assert.NoError(t, err)

	got1, err := repo.GetNotificationByID(ctx, ntf1.ID)
	assert.NoError(t, err)
	assert.Equal(t, ntf1.Text, got1.Text)
	assert.Equal(t, models.StatusInFlight, got1.Status)
	assert.Equal(t, 1, got1.Attempts)

	got2, err := repo.GetNotificationByID(ctx, ntf2.ID)
	assert.NoError(t, err)
	assert.Equal(t, ntf2.RecipientPhone, got2.RecipientPhone)
}

func TestNotificationRepository_GetNotificationByID_NotExists(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewNotificationRepository(testPool)

	_, err := repo.GetNotificationByID(ctx, uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotificationNotExists)
}

func TestNotificationRepository_ChangeNotificationStatus(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewNotificationRepository(testPool)

	ntf := &models.Notification{
		ID:             uuid.New(),
		UserID:         200,
		Text:           "Will update",
		RecipientPhone: "+10000000003",
	}

	err := repo.CreateMultipleNotifications(ctx, []*models.Notification{ntf})
	assert.NoError(t, err)

	err = repo.ChangeNotificationStatus(ctx, ntf.ID, models.StatusSent)
	assert.NoError(t, err)

	updated, err := repo.GetNotificationByID(ctx, ntf.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusSent, updated.Status)
	assert.WithinDuration(t, time.Now(), updated.UpdatedAt, time.Second*2)
}

func TestNotificationRepository_ChangeNotificationStatus_NotExists(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewNotificationRepository(testPool)

	err := repo.ChangeNotificationStatus(ctx, uuid.New(), models.StatusFailed)
	assert.ErrorIs(t, err, domain.ErrNotificationNotExists)
}
