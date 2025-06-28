package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTemplateRepo defines a mock implementation of domain.TemplateRepository
// generated for use in unit tests.
type MockTemplateRepo struct {
	mock.Mock
}

func (m *MockTemplateRepo) GetTemplatesByUserID(ctx context.Context, userID int) ([]*models.Template, error) {
	args := m.Called(ctx, userID)
	if tmpls := args.Get(0); tmpls != nil {
		return tmpls.([]*models.Template), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTemplateRepo) GetTemplateByID(ctx context.Context, userID, tmplID int) (*models.Template, error) {
	args := m.Called(ctx, userID, tmplID)
	if tmpl := args.Get(0); tmpl != nil {
		return tmpl.(*models.Template), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTemplateRepo) CreateTemplate(ctx context.Context, tmpl *models.Template) (*models.Template, error) {
	args := m.Called(ctx, tmpl)
	if t := args.Get(0); t != nil {
		return t.(*models.Template), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTemplateRepo) UpdateTemplate(ctx context.Context, userID, tmplID int, tmpl *models.Template) (*models.Template, error) {
	args := m.Called(ctx, userID, tmplID, tmpl)
	if t := args.Get(0); t != nil {
		return t.(*models.Template), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTemplateRepo) DeleteTemplate(ctx context.Context, userID, tmplID int) error {
	return m.Called(ctx, userID, tmplID).Error(0)
}

func TestTemplateService(t *testing.T) {
	tmplA := &models.Template{ID: 1, UserID: 10, Body: "A"}
	tmplB := &models.Template{ID: 2, UserID: 10, Body: "B"}
	errDB := errors.New("db error")

	tests := []struct {
		name       string
		method     string // GetAll, GetByID, Create, Update, Delete
		args       []interface{}
		mockSetup  func(*MockTemplateRepo)
		wantErr    error
		wantResult interface{}
	}{
		{
			name:   "GetTemplatesByUserID success",
			method: "GetTemplatesByUserID",
			args:   []interface{}{context.Background(), 10},
			mockSetup: func(m *MockTemplateRepo) {
				m.On("GetTemplatesByUserID", mock.Anything, 10).
					Return([]*models.Template{tmplA, tmplB}, nil).
					Once()
			},
			wantErr:    nil,
			wantResult: []*models.Template{tmplA, tmplB},
		},
		{
			name:   "GetTemplateByID not found error",
			method: "GetTemplateByID",
			args:   []interface{}{context.Background(), 10, 99},
			mockSetup: func(m *MockTemplateRepo) {
				m.On("GetTemplateByID", mock.Anything, 10, 99).
					Return(nil, errDB).
					Once()
			},
			wantErr:    errDB,
			wantResult: nil,
		},
		{
			name:       "CreateTemplate invalid body",
			method:     "CreateTemplate",
			args:       []interface{}{context.Background(), &models.Template{UserID: 10, Body: ""}},
			mockSetup:  func(m *MockTemplateRepo) {},
			wantErr:    domain.ErrInvalidTemplate,
			wantResult: nil,
		},
		{
			name:   "CreateTemplate success",
			method: "CreateTemplate",
			args:   []interface{}{context.Background(), &models.Template{UserID: 10, Body: "New", Name: "New"}},
			mockSetup: func(m *MockTemplateRepo) {
				m.On("CreateTemplate", mock.Anything, mock.MatchedBy(func(x *models.Template) bool {
					return x.UserID == 10 && x.Body == "New" && x.Name == "New"
				})).
					Return(&models.Template{ID: 3, UserID: 10, Body: "New"}, nil).
					Once()
			},
			wantErr:    nil,
			wantResult: &models.Template{ID: 3, UserID: 10, Body: "New"},
		},
		{
			name:       "UpdateTemplate invalid body",
			method:     "UpdateTemplate",
			args:       []interface{}{context.Background(), 10, 1, &models.Template{Body: ""}},
			mockSetup:  func(m *MockTemplateRepo) {},
			wantErr:    domain.ErrInvalidTemplate,
			wantResult: nil,
		},
		{
			name:   "UpdateTemplate success",
			method: "UpdateTemplate",
			args:   []interface{}{context.Background(), 10, 1, &models.Template{Body: "Upd", Name: "Upd"}},
			mockSetup: func(m *MockTemplateRepo) {
				m.On("UpdateTemplate", mock.Anything, 10, 1, mock.MatchedBy(func(x *models.Template) bool {
					return x.Body == "Upd" && x.Name == "Upd"
				})).
					Return(&models.Template{ID: 1, UserID: 10, Body: "Upd"}, nil).
					Once()
			},
			wantErr:    nil,
			wantResult: &models.Template{ID: 1, UserID: 10, Body: "Upd"},
		},
		{
			name:   "DeleteTemplate success",
			method: "DeleteTemplate",
			args:   []interface{}{context.Background(), 10, 1},
			mockSetup: func(m *MockTemplateRepo) {
				m.On("DeleteTemplate", mock.Anything, 10, 1).
					Return(nil).
					Once()
			},
			wantErr:    nil,
			wantResult: nil,
		},
		{
			name:   "DeleteTemplate error",
			method: "DeleteTemplate",
			args:   []interface{}{context.Background(), 10, 2},
			mockSetup: func(m *MockTemplateRepo) {
				m.On("DeleteTemplate", mock.Anything, 10, 2).
					Return(errDB).
					Once()
			},
			wantErr:    errDB,
			wantResult: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockTemplateRepo)
			svc := service.NewTemplateService(mockRepo)
			// setup expectations
			tc.mockSetup(mockRepo)

			var (
				res interface{}
				err error
			)

			switch tc.method {
			case "GetTemplatesByUserID":
				res, err = svc.GetTemplatesByUserID(tc.args[0].(context.Context), tc.args[1].(int))
			case "GetTemplateByID":
				res, err = svc.GetTemplateByID(tc.args[0].(context.Context), tc.args[1].(int), tc.args[2].(int))
			case "CreateTemplate":
				res, err = svc.CreateTemplate(tc.args[0].(context.Context), tc.args[1].(*models.Template))
			case "UpdateTemplate":
				res, err = svc.UpdateTemplate(tc.args[0].(context.Context), tc.args[1].(int), tc.args[2].(int), tc.args[3].(*models.Template))
			case "DeleteTemplate":
				err = svc.DeleteTemplate(tc.args[0].(context.Context), tc.args[1].(int), tc.args[2].(int))
			}

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}

			if tc.wantResult != nil {
				assert.Equal(t, tc.wantResult, res)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
