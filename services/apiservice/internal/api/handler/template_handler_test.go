package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- GET /templates ---
func TestTemplateHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		setup      func(m *MockTemplateService)
		wantStatus int
		wantBody   []*models.Template
	}{
		{
			name:   "success",
			userID: 1,
			setup: func(m *MockTemplateService) {
				m.
					On("GetTemplatesByUserID", mock.Anything, 1).
					Return([]*models.Template{{ID: 1, UserID: 1, Name: "T1", Body: "B1"}}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   []*models.Template{{ID: 1, UserID: 1, Name: "T1", Body: "B1"}},
		},
		{
			name:   "service error",
			userID: 2,
			setup: func(m *MockTemplateService) {
				m.
					On("GetTemplatesByUserID", mock.Anything, 2).
					Return(([]*models.Template)(nil), assert.AnError).
					Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockTemplateService)
			tc.setup(m)
			h := handler.NewTemplateHandler(m, logger, timeout)

			req := httptest.NewRequest("GET", "/templates", nil)
			req = injectUserID(req, tc.userID)
			rr := httptest.NewRecorder()

			h.Get(rr, req)
			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			if tc.wantStatus == http.StatusOK {
				var got []*models.Template
				err := json.NewDecoder(res.Body).Decode(&got)
				assert.NoError(t, err)
				assert.Equal(t, tc.wantBody, got)
			}
		})
	}
}

// --- GET /templates/{id} ---
func TestTemplateHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		idParam    string
		setup      func(m *MockTemplateService)
		wantStatus int
		wantBody   *models.Template
	}{
		{
			name:       "bad id",
			userID:     1,
			idParam:    "abc",
			setup:      func(m *MockTemplateService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "not found",
			userID:  1,
			idParam: "10",
			setup: func(m *MockTemplateService) {
				m.
					On("GetTemplateByID", mock.Anything, 1, 10).
					Return((*models.Template)(nil), domain.ErrTemplateNotExists).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "service error",
			userID:  1,
			idParam: "11",
			setup: func(m *MockTemplateService) {
				m.
					On("GetTemplateByID", mock.Anything, 1, 11).
					Return((*models.Template)(nil), assert.AnError).
					Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:    "success",
			userID:  1,
			idParam: "5",
			setup: func(m *MockTemplateService) {
				m.
					On("GetTemplateByID", mock.Anything, 1, 5).
					Return(&models.Template{ID: 5, UserID: 1, Name: "N", Body: "P"}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   &models.Template{ID: 5, UserID: 1, Name: "N", Body: "P"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockTemplateService)
			tc.setup(m)
			h := handler.NewTemplateHandler(m, logger, timeout)

			req := httptest.NewRequest("GET", "/templates/"+tc.idParam, nil)
			req = injectUserID(req, tc.userID)
			req = mux.SetURLVars(req, map[string]string{"id": tc.idParam})
			rr := httptest.NewRecorder()

			h.GetByID(rr, req)
			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			if tc.wantStatus == http.StatusOK {
				var got models.Template
				err := json.NewDecoder(res.Body).Decode(&got)
				assert.NoError(t, err)
				assert.Equal(t, tc.wantBody, &got)
			}
		})
	}
}

// --- POST /templates ---
func TestTemplateHandler_Post(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		body       any
		setup      func(m *MockTemplateService)
		wantStatus int
		wantBody   *models.Template
	}{
		{
			name:       "invalid json",
			userID:     1,
			body:       `{"name":}`,
			setup:      func(m *MockTemplateService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid template",
			userID: 1,
			body:   domain.PostTemplateRequest{Name: "", Body: "B"},
			setup: func(m *MockTemplateService) {
				m.
					On("CreateTemplate", mock.Anything, mock.Anything).
					Return((*models.Template)(nil), domain.ErrInvalidTemplate).
					Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "conflict",
			userID: 1,
			body:   domain.PostTemplateRequest{Name: "N", Body: "B"},
			setup: func(m *MockTemplateService) {
				m.
					On("CreateTemplate", mock.Anything, &models.Template{UserID: 1, Name: "N", Body: "B"}).
					Return((*models.Template)(nil), domain.ErrTemplateAlreadyExists).
					Once()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:   "success",
			userID: 1,
			body:   domain.PostTemplateRequest{Name: "N", Body: "B"},
			setup: func(m *MockTemplateService) {
				m.
					On("CreateTemplate", mock.Anything, &models.Template{UserID: 1, Name: "N", Body: "B"}).
					Return(&models.Template{ID: 7, UserID: 1, Name: "N", Body: "B"}, nil).
					Once()
			},
			wantStatus: http.StatusCreated,
			wantBody:   &models.Template{ID: 7, UserID: 1, Name: "N", Body: "B"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockTemplateService)
			tc.setup(m)
			h := handler.NewTemplateHandler(m, logger, timeout)

			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(tc.body)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest("POST", "/templates", &buf)
			req = injectUserID(req, tc.userID)
			rr := httptest.NewRecorder()

			h.Post(rr, req)
			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			if tc.wantBody != nil {
				var got models.Template
				err := json.NewDecoder(res.Body).Decode(&got)
				assert.NoError(t, err)
				assert.Equal(t, tc.wantBody, &got)
			}
		})
	}
}

// --- PUT /templates/{id} ---
func TestTemplateHandler_Put(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		idParam    string
		body       any
		setup      func(m *MockTemplateService)
		wantStatus int
		wantBody   *models.Template
	}{
		{
			name:       "bad id",
			userID:     1,
			idParam:    "x",
			body:       nil,
			setup:      func(m *MockTemplateService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad json",
			userID:     1,
			idParam:    "1",
			body:       `{"name":`,
			setup:      func(m *MockTemplateService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "invalid template",
			userID:  1,
			idParam: "2",
			body:    domain.PutTemplateRequest{Name: "", Body: "B"},
			setup: func(m *MockTemplateService) {
				m.
					On("UpdateTemplate", mock.Anything, 1, 2, mock.Anything).
					Return((*models.Template)(nil), domain.ErrInvalidTemplate).
					Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "not found",
			userID:  1,
			idParam: "3",
			body:    domain.PutTemplateRequest{Name: "N", Body: "B"},
			setup: func(m *MockTemplateService) {
				m.
					On("UpdateTemplate", mock.Anything, 1, 3, &models.Template{UserID: 1, Name: "N", Body: "B"}).
					Return((*models.Template)(nil), domain.ErrTemplateNotExists).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "conflict",
			userID:  1,
			idParam: "4",
			body:    domain.PutTemplateRequest{Name: "N", Body: "B"},
			setup: func(m *MockTemplateService) {
				m.
					On("UpdateTemplate", mock.Anything, 1, 4, &models.Template{UserID: 1, Name: "N", Body: "B"}).
					Return((*models.Template)(nil), domain.ErrTemplateAlreadyExists).
					Once()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:    "success",
			userID:  1,
			idParam: "5",
			body:    domain.PutTemplateRequest{Name: "N", Body: "B"},
			setup: func(m *MockTemplateService) {
				m.
					On("UpdateTemplate", mock.Anything, 1, 5, &models.Template{UserID: 1, Name: "N", Body: "B"}).
					Return(&models.Template{ID: 5, UserID: 1, Name: "N", Body: "B"}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   &models.Template{ID: 5, UserID: 1, Name: "N", Body: "B"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockTemplateService)
			tc.setup(m)
			h := handler.NewTemplateHandler(m, logger, timeout)

			var buf bytes.Buffer
			if tc.body != nil {
				err := json.NewEncoder(&buf).Encode(tc.body)
				if err != nil {
					t.Fatal(err)
				}
			}

			req := httptest.NewRequest("PUT", "/templates/"+tc.idParam, &buf)
			req = injectUserID(req, tc.userID)
			req = mux.SetURLVars(req, map[string]string{"id": tc.idParam})
			rr := httptest.NewRecorder()

			h.Put(rr, req)
			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			if tc.wantBody != nil {
				var got models.Template
				err := json.NewDecoder(res.Body).Decode(&got)
				assert.NoError(t, err)
				assert.Equal(t, tc.wantBody, &got)
			}
		})
	}
}

// --- DELETE /templates/{id} ---
func TestTemplateHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		idParam    string
		setup      func(m *MockTemplateService)
		wantStatus int
	}{
		{
			name:       "bad id",
			userID:     1,
			idParam:    "x",
			setup:      func(m *MockTemplateService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "not found",
			userID:  2,
			idParam: "4",
			setup: func(m *MockTemplateService) {
				m.
					On("DeleteTemplate", mock.Anything, 2, 4).
					Return(domain.ErrTemplateNotExists).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "service error",
			userID:  3,
			idParam: "5",
			setup: func(m *MockTemplateService) {
				m.
					On("DeleteTemplate", mock.Anything, 3, 5).
					Return(assert.AnError).
					Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:    "success",
			userID:  3,
			idParam: "6",
			setup: func(m *MockTemplateService) {
				m.
					On("DeleteTemplate", mock.Anything, 3, 6).
					Return(nil).
					Once()
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockTemplateService)
			tc.setup(m)
			h := handler.NewTemplateHandler(m, logger, timeout)

			req := httptest.NewRequest("DELETE", "/templates/"+tc.idParam, nil)
			req = injectUserID(req, tc.userID)
			req = mux.SetURLVars(req, map[string]string{"id": tc.idParam})
			rr := httptest.NewRecorder()

			h.Delete(rr, req)
			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, res.StatusCode)
		})
	}
}
