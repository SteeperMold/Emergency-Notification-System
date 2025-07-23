package repository

import (
	"context"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/models"
	"github.com/jackc/pgx/v5"
)

type ContactsRepository struct {
	db domain.DBConn
}

func NewContactsRepository(db domain.DBConn) *ContactsRepository {
	return &ContactsRepository{
		db: db,
	}
}

func (cr *ContactsRepository) SaveContacts(ctx context.Context, contacts []*models.Contact) error {
	tx, err := cr.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	const tempTableQuery = `
		CREATE TEMP TABLE contacts_stage
		(
			user_id INT,
			name    TEXT,
			phone   TEXT
		) ON COMMIT DROP 
	`

	_, err = tx.Exec(ctx, tempTableQuery)
	if err != nil {
		return err
	}

	rows := make([][]any, len(contacts))
	for i, c := range contacts {
		rows[i] = []any{c.UserID, c.Name, c.Phone}
	}

	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"contacts_stage"},
		[]string{"user_id", "name", "phone"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	const insertQuery = `
		INSERT INTO contacts(user_id, name, phone)
		SELECT user_id, name, phone
		FROM contacts_stage
		ON CONFLICT (user_id, phone) DO NOTHING
	`

	_, err = tx.Exec(ctx, insertQuery)
	if err != nil {
		return err
	}

	return nil
}
