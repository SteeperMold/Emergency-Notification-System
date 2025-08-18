package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func injectUserID(r *http.Request, uid int) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextkeys.UserID, uid))
}

var (
	logger  = zap.NewNop()
	timeout = 5 * time.Millisecond
)

// --- GET /contacts ---
func TestContactsHandler_Get(t *testing.T) {
	type resp struct {
		Contacts []*models.Contact `json:"contacts"`
		Total    int               `json:"total"`
	}

	tests := []struct {
		name       string
		userID     int
		setup      func(m *MockContactsService)
		wantStatus int
		wantBody   resp
	}{
		{
			name:   "success",
			userID: 1,
			setup: func(m *MockContactsService) {
				m.
					On("GetContactsPageByUserID", mock.Anything, 1, mock.Anything, mock.Anything).
					Return([]*models.Contact{{ID: 5, UserID: 1, Name: "A", Phone: "P"}}, nil).
					Once()
				m.
					On("GetContactsCountByUserID", mock.Anything, 1).
					Return(1, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: resp{
				Contacts: []*models.Contact{
					{ID: 5, UserID: 1, Name: "A", Phone: "P"},
				},
				Total: 1,
			},
		},
		{
			name:   "service error",
			userID: 2,
			setup: func(m *MockContactsService) {
				m.
					On("GetContactsPageByUserID", mock.Anything, 2, mock.Anything, mock.Anything).
					Return(([]*models.Contact)(nil), assert.AnError).
					Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockContactsService)
			tc.setup(m)
			h := handler.NewContactsHandler(m, logger, timeout)

			req := httptest.NewRequest("GET", "/contacts", nil)
			req = injectUserID(req, tc.userID)
			rr := httptest.NewRecorder()

			h.Get(rr, req)
			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			if tc.wantStatus == http.StatusOK {
				var got resp
				err := json.NewDecoder(res.Body).Decode(&got)
				assert.NoError(t, err)
				assert.Equal(t, tc.wantBody, got)
			}
		})
	}
}

// --- GET /contacts/{id} ---
func TestContactsHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		idParam    string
		setup      func(m *MockContactsService)
		wantStatus int
		wantBody   *models.Contact
	}{
		{
			name:       "bad id",
			userID:     1,
			idParam:    "abc",
			setup:      func(m *MockContactsService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "not found",
			userID:  1,
			idParam: "10",
			setup: func(m *MockContactsService) {
				m.
					On("GetContactByID", mock.Anything, 1, 10).
					Return((*models.Contact)(nil), domain.ErrContactNotExists).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "success",
			userID:  1,
			idParam: "7",
			setup: func(m *MockContactsService) {
				m.
					On("GetContactByID", mock.Anything, 1, 7).
					Return(&models.Contact{ID: 7, UserID: 1, Name: "X", Phone: "P"}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   &models.Contact{ID: 7, UserID: 1, Name: "X", Phone: "P"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockContactsService)
			tc.setup(m)
			h := handler.NewContactsHandler(m, logger, timeout)

			req := httptest.NewRequest("GET", "/contacts/"+tc.idParam, nil)
			req = injectUserID(req, tc.userID)
			// inject route var
			req = mux.SetURLVars(req, map[string]string{"id": tc.idParam})

			rr := httptest.NewRecorder()
			h.GetByID(rr, req)
			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			if tc.wantStatus == http.StatusOK {
				var got models.Contact
				err := json.NewDecoder(res.Body).Decode(&got)
				assert.NoError(t, err)
				assert.Equal(t, tc.wantBody, &got)
			}
		})
	}
}

// --- POST /contacts ---
func TestContactsHandler_Post(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		body       any
		setup      func(m *MockContactsService)
		wantStatus int
		wantBody   *models.Contact
	}{
		{
			name:       "invalid json",
			userID:     1,
			body:       `{"name":}`,
			setup:      func(m *MockContactsService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid payload",
			userID: 1,
			body:   domain.PostContactRequest{Name: "", Phone: "P"},
			setup: func(m *MockContactsService) {
				m.
					On("CreateContact", mock.Anything, mock.Anything).
					Return((*models.Contact)(nil), domain.ErrInvalidContact).
					Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "conflict",
			userID: 1,
			body:   domain.PostContactRequest{Name: "A", Phone: "P"},
			setup: func(m *MockContactsService) {
				m.
					On("CreateContact", mock.Anything, &models.Contact{UserID: 1, Name: "A", Phone: "P"}).
					Return((*models.Contact)(nil), domain.ErrContactAlreadyExists).
					Once()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:   "success",
			userID: 1,
			body:   domain.PostContactRequest{Name: "A", Phone: "P"},
			setup: func(m *MockContactsService) {
				m.
					On("CreateContact", mock.Anything, &models.Contact{UserID: 1, Name: "A", Phone: "P"}).
					Return(&models.Contact{ID: 9, UserID: 1, Name: "A", Phone: "P"}, nil).
					Once()
			},
			wantStatus: http.StatusCreated,
			wantBody:   &models.Contact{ID: 9, UserID: 1, Name: "A", Phone: "P"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockContactsService)
			tc.setup(m)
			h := handler.NewContactsHandler(m, logger, timeout)

			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(tc.body); err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest("POST", "/contacts", &buf)
			req = injectUserID(req, tc.userID)
			rr := httptest.NewRecorder()

			h.Post(rr, req)
			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			if tc.wantBody != nil {
				var got models.Contact
				assert.NoError(t, json.NewDecoder(res.Body).Decode(&got))
				assert.Equal(t, tc.wantBody, &got)
			}
		})
	}
}

// --- PUT /contacts/{id} ---
func TestContactsHandler_Put(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		idParam    string
		body       any
		setup      func(m *MockContactsService)
		wantStatus int
		wantBody   *models.Contact
	}{
		{
			name:       "bad id",
			userID:     1,
			idParam:    "x",
			body:       nil,
			setup:      func(m *MockContactsService) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name:       "bad json",
			userID:     1,
			idParam:    "1",
			body:       `{"name":`,
			setup:      func(m *MockContactsService) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   nil,
		},
		{
			name:    "not found",
			userID:  1,
			idParam: "2",
			body:    domain.PutContactRequest{Name: "N", Phone: "P"},
			setup: func(m *MockContactsService) {
				m.
					On("UpdateContact", mock.Anything, 1, 2, &models.Contact{UserID: 1, Name: "N", Phone: "P"}).
					Return((*models.Contact)(nil), domain.ErrContactNotExists).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "success",
			userID:  1,
			idParam: "3",
			body:    domain.PutContactRequest{Name: "N", Phone: "P"},
			setup: func(m *MockContactsService) {
				m.
					On("UpdateContact", mock.Anything, 1, 3, &models.Contact{UserID: 1, Name: "N", Phone: "P"}).
					Return(&models.Contact{ID: 3, UserID: 1, Name: "N", Phone: "P"}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   &models.Contact{ID: 3, UserID: 1, Name: "N", Phone: "P"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockContactsService)
			tc.setup(m)
			h := handler.NewContactsHandler(m, logger, timeout)

			var buf bytes.Buffer
			if tc.body != nil {
				_ = json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest("PUT", "/contacts/"+tc.idParam, &buf)
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
				var got models.Contact
				assert.NoError(t, json.NewDecoder(res.Body).Decode(&got))
				assert.Equal(t, tc.wantBody, &got)
			}
		})
	}
}

// --- DELETE /contacts/{id} ---
func TestContactsHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		idParam    string
		setup      func(m *MockContactsService)
		wantStatus int
	}{
		{
			name:       "bad id",
			userID:     1,
			idParam:    "y",
			setup:      func(m *MockContactsService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "not found",
			userID:  2,
			idParam: "4",
			setup: func(m *MockContactsService) {
				m.
					On("DeleteContact", mock.Anything, 2, 4).
					Return(domain.ErrContactNotExists).
					Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "success",
			userID:  2,
			idParam: "5",
			setup: func(m *MockContactsService) {
				m.
					On("DeleteContact", mock.Anything, 2, 5).
					Return(nil).
					Once()
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockContactsService)
			tc.setup(m)
			h := handler.NewContactsHandler(m, logger, timeout)

			req := httptest.NewRequest("DELETE", "/contacts/"+tc.idParam, nil)
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
