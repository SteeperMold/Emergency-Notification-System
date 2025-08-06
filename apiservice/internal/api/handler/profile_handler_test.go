package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestProfileHandler_GetProfile(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*http.Request)
		setupMock      func(m *MockProfileService)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "user ID missing from context",
			setupContext:   func(r *http.Request) {},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "internal server error\n",
		},
		{
			name: "user not found",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 101)
				*r = *r.WithContext(ctx)
			},
			setupMock: func(m *MockProfileService) {
				m.On("GetUserByID", mock.Anything, 101).
					Return((*models.User)(nil), domain.ErrUserNotExists).
					Once()
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       "user not exists\n",
		},
		{
			name: "internal error from service",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 102)
				*r = *r.WithContext(ctx)
			},
			setupMock: func(m *MockProfileService) {
				m.On("GetUserByID", mock.Anything, 102).
					Return((*models.User)(nil), assert.AnError).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "internal server error\n",
		},
		{
			name: "successful profile fetch",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 103)
				*r = *r.WithContext(ctx)
			},
			setupMock: func(m *MockProfileService) {
				m.On("GetUserByID", mock.Anything, 103).
					Return(&models.User{ID: 103, Email: "jane@example.com"}, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"id":103,"email":"jane@example.com","creationTime":"0001-01-01T00:00:00Z"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := new(MockProfileService)

			if tc.setupMock != nil {
				tc.setupMock(mockSvc)
			}

			h := handler.NewProfileHandler(mockSvc, zap.NewNop(), 100*time.Millisecond)

			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			if tc.setupContext != nil {
				tc.setupContext(req)
			}

			rec := httptest.NewRecorder()
			h.GetProfile(rec, req)

			require.Equal(t, tc.wantStatusCode, rec.Code)

			if tc.wantStatusCode == http.StatusOK {
				require.JSONEq(t, tc.wantBody, rec.Body.String())
			} else {
				require.Equal(t, tc.wantBody, rec.Body.String())
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
