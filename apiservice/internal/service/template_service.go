package service

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
)

// TemplateService provides operations for managing message templates.
type TemplateService struct {
	repository domain.TemplateRepository
}

// NewTemplateService creates and returns a new TemplateService with the given template repository.
func NewTemplateService(r domain.TemplateRepository) *TemplateService {
	return &TemplateService{
		repository: r,
	}
}

// GetTemplatesByUserID retrieves all templates belonging to the specified user.
func (ts *TemplateService) GetTemplatesByUserID(ctx context.Context, userID int) ([]*models.Template, error) {
	return ts.repository.GetTemplatesByUserID(ctx, userID)
}

// GetTemplateByID retrieves a specific template by its ID for the given user.
func (ts *TemplateService) GetTemplateByID(ctx context.Context, userID int, tmplID int) (*models.Template, error) {
	return ts.repository.GetTemplateByID(ctx, userID, tmplID)
}

// CreateTemplate validates and creates a new message template.
// Returns the created Template model or a domain.ErrInvalidTemplate if body length is invalid.
func (ts *TemplateService) CreateTemplate(ctx context.Context, tmpl *models.Template) (*models.Template, error) {
	if len(tmpl.Body) == 0 || len(tmpl.Body) > 256 {
		return nil, domain.ErrInvalidTemplate
	}

	return ts.repository.CreateTemplate(ctx, tmpl)
}

// UpdateTemplate validates and updates an existing message template for the user.
// Returns the updated Template model or a domain.ErrInvalidTemplate / domain.ErrTemplateNotExists as appropriate.
func (ts *TemplateService) UpdateTemplate(ctx context.Context, userID int, tmplID int, updatedTmpl *models.Template) (*models.Template, error) {
	if len(updatedTmpl.Body) == 0 || len(updatedTmpl.Body) > 256 {
		return nil, domain.ErrInvalidTemplate
	}

	return ts.repository.UpdateTemplate(ctx, userID, tmplID, updatedTmpl)
}

// DeleteTemplate removes the specified template for the user.
// Returns domain.ErrTemplateNotExists if no rows were deleted.
func (ts *TemplateService) DeleteTemplate(ctx context.Context, userID, tmplID int) error {
	return ts.repository.DeleteTemplate(ctx, userID, tmplID)
}
