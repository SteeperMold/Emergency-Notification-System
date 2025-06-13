package repository

import (
	"context"
	"errors"

	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
	"github.com/jackc/pgx/v5"
)

// TemplateRepository handles CRUD operations on the message_templates table.
type TemplateRepository struct {
	db domain.DBConn
}

// NewTemplateRepository constructs a TemplateRepository using the provided DB connection.
func NewTemplateRepository(db domain.DBConn) *TemplateRepository {
	return &TemplateRepository{
		db: db,
	}
}

// GetTemplatesByUserID retrieves all templates for a given user.
// Returns an error if the query fails.
func (tr *TemplateRepository) GetTemplatesByUserID(ctx context.Context, userID int) ([]*models.Template, error) {
	const q = `
		SELECT id, user_id, body, created_at, updated_at
		FROM message_templates
		WHERE user_id = $1
	`

	templates := make([]*models.Template, 0)

	rows, err := tr.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Template

		err := rows.Scan(&t.ID, &t.UserID, &t.Body, &t.CreationTime, &t.UpdateTime)
		if err != nil {
			return nil, err
		}

		templates = append(templates, &t)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return templates, nil
}

// GetTemplateByID retrieves a single template by user ID and template ID.
// Returns domain.ErrTemplateNotExists if no matching row is found.
func (tr *TemplateRepository) GetTemplateByID(ctx context.Context, userID int, tmplID int) (*models.Template, error) {
	const q = `
		SELECT id, user_id, body, created_at, updated_at
		FROM message_templates
		WHERE user_id = $1
		  AND id = $2
	`

	var tmpl models.Template

	row := tr.db.QueryRow(ctx, q, userID, tmplID)
	err := row.Scan(&tmpl.ID, &tmpl.UserID, &tmpl.Body, &tmpl.CreationTime, &tmpl.UpdateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotExists
		}

		return nil, err
	}

	return &tmpl, nil
}

// CreateTemplate inserts a new template and returns the created record.
// Returns an error if insertion fails.
func (tr *TemplateRepository) CreateTemplate(ctx context.Context, tmpl *models.Template) (*models.Template, error) {
	const q = `
		INSERT INTO message_templates (user_id, body)
		VALUES ($1, $2)
		RETURNING id, user_id, body, created_at, updated_at
	`

	var newTmpl models.Template

	row := tr.db.QueryRow(ctx, q, tmpl.UserID, tmpl.Body)
	err := row.Scan(&newTmpl.ID, &newTmpl.UserID, &newTmpl.Body, &newTmpl.CreationTime, &newTmpl.UpdateTime)
	if err != nil {
		return nil, err
	}

	return &newTmpl, nil
}

// UpdateTemplate modifies an existing template’s body and updated_at timestamp.
// Returns domain.ErrTemplateNotExists if no template was updated.
func (tr *TemplateRepository) UpdateTemplate(ctx context.Context, userID int, tmplID int, updatedTmpl *models.Template) (*models.Template, error) {
	const q = `
		UPDATE message_templates
		SET user_id    = $1,
			body       = $2,
			updated_at = now()
		WHERE id = $3
		  AND user_id = $4
		RETURNING id, user_id, body, created_at, updated_at;
	`

	row := tr.db.QueryRow(ctx, q, updatedTmpl.UserID, updatedTmpl.Body, tmplID, userID)

	var t models.Template
	err := row.Scan(&t.ID, &t.UserID, &t.Body, &t.CreationTime, &t.UpdateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotExists
		}

		return nil, err
	}

	return &t, nil
}

// DeleteTemplate removes a template by ID and user ID.
// Returns domain.ErrTemplateNotExists if no row was deleted.
func (tr *TemplateRepository) DeleteTemplate(ctx context.Context, userID int, tmplID int) error {
	const q = `
		DELETE
		FROM message_templates
		WHERE id = $1
		  AND user_id = $2;
	`

	res, err := tr.db.Exec(ctx, q, tmplID, userID)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return domain.ErrTemplateNotExists
	}

	return nil
}
