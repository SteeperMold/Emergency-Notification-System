package repository

import (
	"context"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/rebalancer-service/internal/models"
)

type NotificationRepository struct {
	db domain.DBConn
}

func NewNotificationRepository(db domain.DBConn) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

func (nr *NotificationRepository) FetchAndUpdatePending(ctx context.Context, limit int) ([]*models.Notification, error) {
	const q = `
		WITH to_dequeue AS (
			SELECT id
			FROM notifications
			WHERE (status = 'pending' AND next_run_at <= now())
			   OR (status = 'in_flight' AND updated_at <= now() - interval '5 minute')
			ORDER BY next_run_at
			LIMIT $1 FOR UPDATE SKIP LOCKED
		)
		UPDATE notifications n
		SET status     = 'in_flight',
			attempts   = attempts + 1,
			updated_at = now()
		FROM to_dequeue d
		WHERE n.id = d.id
		RETURNING n.id, n.user_id, n.text, n.recipient_phone, n.status, n.attempts, n.next_run_at, n.created_at, n.updated_at
	`

	rows, err := nr.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification

	for rows.Next() {
		var n models.Notification

		err := rows.Scan(&n.ID, &n.UserID, &n.Text, &n.RecipientPhone, &n.Status, &n.Attempts, &n.NextRunAt, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, &n)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return notifications, nil
}
