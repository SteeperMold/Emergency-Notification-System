package repository

import (
	"context"
	"errors"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ContactsRepository handles CRUD operations on the contacts table.
type ContactsRepository struct {
	db domain.DBConn
}

// NewContactsRepository constructs a ContactsRepository backed by the given DB connection.
func NewContactsRepository(db domain.DBConn) *ContactsRepository {
	return &ContactsRepository{
		db: db,
	}
}

// GetAllContactsByUserID retrieves all contacts for a specific user identified by userID.
func (cr *ContactsRepository) GetAllContactsByUserID(ctx context.Context, userID int) ([]*models.Contact, error) {
	const q = `
		SELECT id, user_id, name, phone, created_at, updated_at
		FROM contacts
		WHERE user_id = $1
	`

	contacts := make([]*models.Contact, 0)

	rows, err := cr.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c models.Contact

		err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.CreationTime, &c.UpdateTime)
		if err != nil {
			return nil, err
		}

		contacts = append(contacts, &c)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

// GetContactsByUserID retrieves a paginated list of contacts for the specified user.
// It applies the given limit and offset for pagination.
func (cr *ContactsRepository) GetContactsByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Contact, error) {
	const q = `
		SELECT id, user_id, name, phone, created_at, updated_at
		FROM contacts
		WHERE user_id = $1
		ORDER BY id
		LIMIT $2 OFFSET $3
	`

	contacts := make([]*models.Contact, 0)

	rows, err := cr.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c models.Contact

		err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.CreationTime, &c.UpdateTime)
		if err != nil {
			return nil, err
		}

		contacts = append(contacts, &c)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

// GetContactByID retrieves a single contact for a user by its contact ID.
// Returns domain.ErrContactNotExists if no row is found.
func (cr *ContactsRepository) GetContactByID(ctx context.Context, userID int, contactID int) (*models.Contact, error) {
	const q = `
		SELECT id, user_id, name, phone, created_at, updated_at
		FROM contacts
		WHERE user_id = $1
		  AND id = $2
	`

	var c models.Contact

	row := cr.db.QueryRow(ctx, q, userID, contactID)
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.CreationTime, &c.UpdateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContactNotExists
		}

		return nil, err
	}

	return &c, nil
}

// CreateContact inserts a new contact and returns the created record.
// If the unique constraint on (user_id, name, phone) is violated, returns domain.ErrContactAlreadyExists.
func (cr *ContactsRepository) CreateContact(ctx context.Context, contact *models.Contact) (*models.Contact, error) {
	const q = `
		INSERT INTO contacts (user_id, name, phone)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, phone, created_at, updated_at
	`

	var c models.Contact

	row := cr.db.QueryRow(ctx, q, contact.UserID, contact.Name, contact.Phone)
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.CreationTime, &c.UpdateTime)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrContactAlreadyExists
		}

		return nil, err
	}

	return &c, nil
}

// UpdateContact modifies an existing contact's name and phone, updating its timestamp.
// Returns domain.ErrContactNotExists if no row matches, or domain.ErrContactAlreadyExists on unique violation.
func (cr *ContactsRepository) UpdateContact(ctx context.Context, userID int, contactID int, updatedContact *models.Contact) (*models.Contact, error) {
	const q = `
		UPDATE contacts
		SET user_id    = $1,
			name       = $2,
			phone      = $3,
			updated_at = now()
		WHERE id = $4
		  AND user_id = $5
		RETURNING id, user_id, name, phone, created_at, updated_at
	`

	row := cr.db.QueryRow(ctx, q, updatedContact.UserID, updatedContact.Name, updatedContact.Phone, contactID, userID)

	var c models.Contact
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.CreationTime, &c.UpdateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContactNotExists
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrContactAlreadyExists
		}

		return nil, err
	}

	return &c, nil
}

// DeleteContact removes a contact record for a user. Returns domain.ErrContactNotExists if no deletion occurred.
func (cr *ContactsRepository) DeleteContact(ctx context.Context, userID int, contactID int) error {
	const q = `
		DELETE
		FROM contacts
		WHERE id = $1
		  AND user_id = $2
	`

	res, err := cr.db.Exec(ctx, q, contactID, userID)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return domain.ErrContactNotExists
	}

	return nil
}
