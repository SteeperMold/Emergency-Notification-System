package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTemplateService_GetTemplatesByUserID(t *testing.T) {
	expected := []*models.Template{
		{ID: 1, UserID: 42, Name: "T1", Body: "B1"},
	}
	m := new(MockTemplateRepository)
	m.
		On("GetTemplatesByUserID", mock.Anything, 42, mock.Anything, mock.Anything).
		Return(expected, nil).
		Once()

	svc := service.NewTemplateService(m, 50, 100)
	out, err := svc.GetTemplatesByUserID(context.Background(), 42, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, expected, out)
	m.AssertExpectations(t)
}

func TestTemplateService_GetTemplateByID(t *testing.T) {
	expected := &models.Template{ID: 2, UserID: 42, Name: "T2", Body: "B2"}
	m := new(MockTemplateRepository)
	m.
		On("GetTemplateByID", mock.Anything, 42, 2).
		Return(expected, nil).
		Once()

	svc := service.NewTemplateService(m, 50, 100)
	out, err := svc.GetTemplateByID(context.Background(), 42, 2)

	assert.NoError(t, err)
	assert.Equal(t, expected, out)
	m.AssertExpectations(t)
}

func TestTemplateService_CreateTemplate(t *testing.T) {
	type args struct {
		tmpl *models.Template
	}

	tests := []struct {
		name      string
		args      args
		mockSetup func(m *MockTemplateRepository)
		want      *models.Template
		wantErr   error
	}{
		{
			name:    "empty name",
			args:    args{tmpl: &models.Template{UserID: 1, Name: "", Body: "ok"}},
			wantErr: domain.ErrInvalidTemplate,
		},
		{
			name:    "name too long",
			args:    args{tmpl: &models.Template{UserID: 1, Name: strings.Repeat("x", 33), Body: "ok"}},
			wantErr: domain.ErrInvalidTemplate,
		},
		{
			name:    "empty body",
			args:    args{tmpl: &models.Template{UserID: 1, Name: "n", Body: ""}},
			wantErr: domain.ErrInvalidTemplate,
		},
		{
			name:    "body too long",
			args:    args{tmpl: &models.Template{UserID: 1, Name: "n", Body: strings.Repeat("y", 257)}},
			wantErr: domain.ErrInvalidTemplate,
		},
		{
			name: "repo error",
			args: args{tmpl: &models.Template{UserID: 1, Name: "n", Body: "b"}},
			mockSetup: func(m *MockTemplateRepository) {
				m.
					On("CreateTemplate", mock.Anything, &models.Template{UserID: 1, Name: "n", Body: "b"}).
					Return((*models.Template)(nil), assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: args{tmpl: &models.Template{UserID: 1, Name: "n", Body: "b"}},
			mockSetup: func(m *MockTemplateRepository) {
				out := &models.Template{ID: 99, UserID: 1, Name: "n", Body: "b"}
				m.
					On("CreateTemplate", mock.Anything, &models.Template{UserID: 1, Name: "n", Body: "b"}).
					Return(out, nil).
					Once()
			},
			want: &models.Template{ID: 99, UserID: 1, Name: "n", Body: "b"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockTemplateRepository)
			if tc.mockSetup != nil {
				tc.mockSetup(m)
			}
			svc := service.NewTemplateService(m, 50, 100)

			out, err := svc.CreateTemplate(context.Background(), tc.args.tmpl)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, out)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, out)
			}
			m.AssertExpectations(t)
		})
	}
}

func TestTemplateService_UpdateTemplate(t *testing.T) {
	type args struct {
		userID int
		tmplID int
		update *models.Template
	}

	tests := []struct {
		name      string
		args      args
		mockSetup func(m *MockTemplateRepository)
		want      *models.Template
		wantErr   error
	}{
		{
			name:    "invalid name",
			args:    args{userID: 1, tmplID: 2, update: &models.Template{UserID: 1, Name: "", Body: "b"}},
			wantErr: domain.ErrInvalidTemplate,
		},
		{
			name:    "invalid body",
			args:    args{userID: 1, tmplID: 2, update: &models.Template{UserID: 1, Name: "n", Body: ""}},
			wantErr: domain.ErrInvalidTemplate,
		},
		{
			name: "repo error",
			args: args{userID: 1, tmplID: 2, update: &models.Template{UserID: 1, Name: "n", Body: "b"}},
			mockSetup: func(m *MockTemplateRepository) {
				m.
					On("UpdateTemplate", mock.Anything, 1, 2, &models.Template{UserID: 1, Name: "n", Body: "b"}).
					Return((*models.Template)(nil), assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: args{userID: 1, tmplID: 2, update: &models.Template{UserID: 1, Name: "n", Body: "b"}},
			mockSetup: func(m *MockTemplateRepository) {
				out := &models.Template{ID: 2, UserID: 1, Name: "n", Body: "b"}
				m.
					On("UpdateTemplate", mock.Anything, 1, 2, &models.Template{UserID: 1, Name: "n", Body: "b"}).
					Return(out, nil).
					Once()
			},
			want: &models.Template{ID: 2, UserID: 1, Name: "n", Body: "b"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockTemplateRepository)
			if tc.mockSetup != nil {
				tc.mockSetup(m)
			}
			svc := service.NewTemplateService(m, 50, 100)

			out, err := svc.UpdateTemplate(context.Background(), tc.args.userID, tc.args.tmplID, tc.args.update)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, out)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, out)
			}

			m.AssertExpectations(t)
		})
	}
}

func TestTemplateService_DeleteTemplate(t *testing.T) {
	m := new(MockTemplateRepository)
	m.
		On("DeleteTemplate", mock.Anything, 1, 2).
		Return(nil).
		Once()

	svc := service.NewTemplateService(m, 50, 100)
	err := svc.DeleteTemplate(context.Background(), 1, 2)

	assert.NoError(t, err)
	m.AssertExpectations(t)
}
