package handler_test

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func makeMultipartRequest(t *testing.T, fieldName, fileName string, content []byte) *http.Request {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(fieldName, fileName)
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/load-contacts", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestLoadContactsHandler_LoadContactsFile(t *testing.T) {
	logger := zap.NewNop()
	timeout := 50 * time.Millisecond

	tests := []struct {
		name           string
		setupContext   func(*http.Request)
		buildRequest   func() *http.Request
		setupMock      func(s *MockLoadContactsService)
		wantStatusCode int
	}{
		{
			name:         "missing userID in context",
			setupContext: func(r *http.Request) {},
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/", nil)
			},
			setupMock:      func(s *MockLoadContactsService) {},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "malformed multipart",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 1)
				*r = *r.WithContext(ctx)
				r.Header.Set("Content-Type", "not/form-data")
			},
			buildRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/", strings.NewReader("oops"))
			},
			setupMock:      func(s *MockLoadContactsService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "missing file field",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 1)
				*r = *r.WithContext(ctx)
			},
			buildRequest: func() *http.Request {
				var buf bytes.Buffer
				w := multipart.NewWriter(&buf)
				_ = w.WriteField("other", "value")
				_ = w.Close()
				req := httptest.NewRequest(http.MethodPost, "/", &buf)
				req.Header.Set("Content-Type", w.FormDataContentType())
				return req
			},
			setupMock:      func(s *MockLoadContactsService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid extension",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 1)
				*r = *r.WithContext(ctx)
			},
			buildRequest: func() *http.Request {
				return makeMultipartRequest(t, "file", "data.txt", []byte("a,b,c\n1,2,3"))
			},
			setupMock:      func(s *MockLoadContactsService) {},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid content type",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 1)
				*r = *r.WithContext(ctx)
			},
			buildRequest: func() *http.Request {
				// valid .csv extension but content type detection returns something else
				return makeMultipartRequest(t, "file", "data.csv", []byte{0x00, 0x01, 0x02})
			},
			setupMock:      func(s *MockLoadContactsService) {},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "service error",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 42)
				*r = *r.WithContext(ctx)
			},
			buildRequest: func() *http.Request {
				return makeMultipartRequest(t, "file", "data.csv", []byte("a,b,c\n1,2,3"))
			},
			setupMock: func(m *MockLoadContactsService) {
				m.
					On("ProcessUpload", mock.Anything, 42, "data.csv", mock.Anything).
					Return(errors.New("oops")).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "success .csv",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 7)
				*r = *r.WithContext(ctx)
			},
			buildRequest: func() *http.Request {
				return makeMultipartRequest(t, "file", "data.csv", []byte("a,b,c\n1,2,3"))
			},
			setupMock: func(m *MockLoadContactsService) {
				m.
					On("ProcessUpload", mock.Anything, 7, "data.csv", mock.Anything).
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusAccepted,
		},
		{
			name: "success .xlsx (ZIP content)",
			setupContext: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), contextkeys.UserID, 99)
				*r = *r.WithContext(ctx)
			},
			buildRequest: func() *http.Request {
				// pretend XLSX is ZIP by giving a ZIP file header
				content := append([]byte("PK\x03\x04"), []byte("rest...")...)
				return makeMultipartRequest(t, "file", "sheet.xlsx", content)
			},
			setupMock: func(m *MockLoadContactsService) {
				m.
					On("ProcessUpload", mock.Anything, 99, "sheet.xlsx", mock.Anything).
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusAccepted,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(MockLoadContactsService)
			tc.setupMock(m)
			h := handler.NewLoadContactsHandler(m, logger, timeout)

			req := tc.buildRequest()
			tc.setupContext(req)
			rr := httptest.NewRecorder()

			h.LoadContactsFile(rr, req)

			m.AssertExpectations(t)

			require.Equal(t, tc.wantStatusCode, rr.Code)
		})
	}
}
