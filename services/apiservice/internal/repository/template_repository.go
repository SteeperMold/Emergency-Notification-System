package repository

import (
	"context"
	"errors"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// GetTemplatesCountByUserID retrieves count of templates belonging to the specified user.
func (tr *TemplateRepository) GetTemplatesCountByUserID(ctx context.Context, userID int) (int, error) {
	const q = `
		SELECT COUNT(*)
		FROM message_templates
		WHERE user_id = $1
	`

	var count int
	err := tr.db.QueryRow(ctx, q, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetTemplatesPageByUserID retrieves a paginated list of templates for the specified user.
// It applies the given limit and offset for pagination.
func (tr *TemplateRepository) GetTemplatesPageByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Template, error) {
	const q = `
		SELECT id, user_id, name, body, created_at, updated_at
		FROM message_templates
		WHERE user_id = $1
		ORDER BY id
		LIMIT $2 OFFSET $3
	`

	templates := make([]*models.Template, 0)

	rows, err := tr.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Template

		err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Body, &t.CreationTime, &t.UpdateTime)
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
		SELECT id, user_id, name, body, created_at, updated_at
		FROM message_templates
		WHERE user_id = $1
		  AND id = $2
	`

	var t models.Template

	row := tr.db.QueryRow(ctx, q, userID, tmplID)
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Body, &t.CreationTime, &t.UpdateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotExists
		}

		return nil, err
	}

	return &t, nil
}

// CreateTemplate inserts a new template and returns the created record.
// Returns an error if insertion fails.
func (tr *TemplateRepository) CreateTemplate(ctx context.Context, tmpl *models.Template) (*models.Template, error) {
	const q = `
		INSERT INTO message_templates (user_id, name, body)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, body, created_at, updated_at
	`

	var t models.Template

	row := tr.db.QueryRow(ctx, q, tmpl.UserID, tmpl.Name, tmpl.Body)
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Body, &t.CreationTime, &t.UpdateTime)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrTemplateAlreadyExists
		}

		return nil, err
	}

	return &t, nil
}

// UpdateTemplate modifies an existing template’s body and updated_at timestamp.
// Returns domain.ErrTemplateNotExists if no template was updated.
func (tr *TemplateRepository) UpdateTemplate(ctx context.Context, userID int, tmplID int, updatedTmpl *models.Template) (*models.Template, error) {
	const q = `
		UPDATE message_templates
		SET user_id    = $1,
		    name       = $2,
			body       = $3,
			updated_at = now()
		WHERE id = $4
		  AND user_id = $5
		RETURNING id, user_id, name, body, created_at, updated_at;
	`

	row := tr.db.QueryRow(ctx, q, updatedTmpl.UserID, updatedTmpl.Name, updatedTmpl.Body, tmplID, userID)

	var t models.Template
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Body, &t.CreationTime, &t.UpdateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotExists
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrTemplateAlreadyExists
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
