package domain

import (
	"context"
	"fmt"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
)

var (
	// ErrTemplateNotExists is returned when a template is not found in the database.
	ErrTemplateNotExists = fmt.Errorf("template doesn't exist")
	// ErrInvalidTemplate is returned when a template body is either too short or exceeds the allowed length.
	ErrInvalidTemplate = fmt.Errorf("template is too long or too short")
	// ErrTemplateAlreadyExists is returned when template with given name already exists
	ErrTemplateAlreadyExists = fmt.Errorf("template already exists")
)

// TemplateRepository defines the interface for persisting and retrieving message templates from a data store.
type TemplateRepository interface {
	GetTemplatesByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Template, error)
	GetTemplateByID(ctx context.Context, userID, templateID int) (*models.Template, error)
	CreateTemplate(ctx context.Context, tmpl *models.Template) (*models.Template, error)
	UpdateTemplate(ctx context.Context, userID, tmplID int, updatedTmpl *models.Template) (*models.Template, error)
	DeleteTemplate(ctx context.Context, userID, tmplID int) error
}

// TemplateService defines the interface for business logic operations on message templates.
type TemplateService interface {
	GetTemplatesByUserID(ctx context.Context, userID, limit, offset int) ([]*models.Template, error)
	GetTemplateByID(ctx context.Context, userID, templateID int) (*models.Template, error)
	CreateTemplate(ctx context.Context, template *models.Template) (*models.Template, error)
	UpdateTemplate(ctx context.Context, userID, tmplID int, updatedTmpl *models.Template) (*models.Template, error)
	DeleteTemplate(ctx context.Context, userID, tmplID int) error
}

// PostTemplateRequest represents the request payload for creating a new template.
type PostTemplateRequest struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// PutTemplateRequest represents the request payload for updating an existing template.
type PutTemplateRequest struct {
	Name string `json:"name"`
	Body string `json:"body"`
}
