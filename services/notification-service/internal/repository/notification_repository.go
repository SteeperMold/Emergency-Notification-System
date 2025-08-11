package repository

import (
	"context"
	"errors"

	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/notification-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// NotificationRepository provides methods to interact with the notifications table in the database.
// It allows batch creation of notifications, retrieval by ID, and status updates.
type NotificationRepository struct {
	db domain.DBConn
}

// NewNotificationRepository constructs a new NotificationRepository
func NewNotificationRepository(db domain.DBConn) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

// CreateMultipleNotifications inserts multiple notification records in a single batch using COPY FROM.
// Each notification is initialized with status "in_flight" and attempts = 1.
func (nr *NotificationRepository) CreateMultipleNotifications(ctx context.Context, notifications []*models.Notification) error {
	rows := make([][]any, len(notifications))
	for i, n := range notifications {
		rows[i] = []any{
			n.ID, n.UserID, n.Text, n.RecipientPhone, "in_flight", 1,
		}
	}

	_, err := nr.db.CopyFrom(ctx, pgx.Identifier{"notifications"}, []string{
		"id", "user_id", "text", "recipient_phone", "status", "attempts",
	}, pgx.CopyFromRows(rows))
	if err != nil {
		return err
	}

	return nil
}

// GetNotificationByID fetches a single notification record by its UUID.
// Returns domain.ErrNotificationNotExists if no matching record is found.
func (nr *NotificationRepository) GetNotificationByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	const q = `
		SELECT id, user_id, text, recipient_phone, status, attempts, next_run_at, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	var n models.Notification

	row := nr.db.QueryRow(ctx, q, id)
	err := row.Scan(&n.ID, &n.UserID, &n.Text, &n.RecipientPhone, &n.Status, &n.Attempts, &n.NextRunAt, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotificationNotExists
		}
		return nil, err
	}

	return &n, nil
}

// ChangeNotificationStatus updates the status and updated_at timestamp of a notification.
// Returns domain.ErrNotificationNotExists if the record does not exist.
func (nr *NotificationRepository) ChangeNotificationStatus(ctx context.Context, id uuid.UUID, newStatus models.NotificationStatus) error {
	const q = `
		UPDATE notifications
		SET status = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	cmdTag, err := nr.db.Exec(ctx, q, id, newStatus)

	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotificationNotExists
	}

	return nil
}
