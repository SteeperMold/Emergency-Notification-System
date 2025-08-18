package repository

import (
	"context"
	"errors"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// NotificationTasksRepository provides access to notification task records in the database.
type NotificationTasksRepository struct {
	db domain.DBConn
}

// NewNotificationTasksRepository creates a new instance of NotificationTasksRepository.
func NewNotificationTasksRepository(db domain.DBConn) *NotificationTasksRepository {
	return &NotificationTasksRepository{
		db: db,
	}
}

// GetNotificationByID retrieves a notification from the database by its UUID.
// Returns domain.ErrNotificationNotExists if the record is not found.
func (ntr *NotificationTasksRepository) GetNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	const q = `
		SELECT id, user_id, recipient_phone, status, attempts, next_run_at, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	row := ntr.db.QueryRow(ctx, q, id)

	var n models.Notification
	err := row.Scan(&n.ID, &n.UserID, &n.RecipientPhone, &n.Status, &n.Attempts, &n.NextRunAt, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotificationNotExists
		}
		return nil, err
	}

	return &n, nil
}

// Reschedule updates a notification task's status to "pending" and sets a new next_run_at timestamp.
// Returns domain.ErrNotificationNotExists if the task doesn't exist.
func (ntr *NotificationTasksRepository) Reschedule(ctx context.Context, id uuid.UUID, nextRunAt time.Time) (*models.Notification, error) {
	const q = `
		UPDATE notifications
		SET status      = 'pending',
		    next_run_at = $2,
			updated_at  = NOW()
		WHERE id = $1
		RETURNING id, user_id, recipient_phone, status, attempts, next_run_at, created_at, updated_at
	`

	row := ntr.db.QueryRow(ctx, q, id, nextRunAt)

	var n models.Notification
	err := row.Scan(&n.ID, &n.UserID, &n.RecipientPhone, &n.Status, &n.Attempts, &n.NextRunAt, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotificationNotExists
		}
		return nil, err
	}

	return &n, nil
}

// MarkFailed marks a notification task as permanently failed by setting its status to "failed".
// Returns domain.ErrNotificationNotExists if the task is not found.
func (ntr *NotificationTasksRepository) MarkFailed(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	const q = `
		UPDATE notifications
		SET status     = 'failed',
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, recipient_phone, status, attempts, next_run_at, created_at, updated_at
	`

	row := ntr.db.QueryRow(ctx, q, id)

	var n models.Notification
	err := row.Scan(&n.ID, &n.UserID, &n.RecipientPhone, &n.Status, &n.Attempts, &n.NextRunAt, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotificationNotExists
		}
		return nil, err
	}

	return &n, nil
}
