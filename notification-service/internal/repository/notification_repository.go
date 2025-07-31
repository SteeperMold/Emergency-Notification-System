package repository

import (
	"context"
	"errors"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NotificationRepository struct {
	db domain.DBConn
}

func NewNotificationRepository(db domain.DBConn) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

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

func (nr *NotificationRepository) ChangeNotificationStatus(ctx context.Context, id uuid.UUID, newStatus models.NotificationStatus) error {
	const q = `
		UPDATE notifications
		SET status = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := nr.db.Exec(ctx, q, id, newStatus)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotificationNotExists
		}
		return err
	}

	return nil
}
