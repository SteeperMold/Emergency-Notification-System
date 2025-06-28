package handler_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/api/handler"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"go.uber.org/zap"
)

type MockLoadContactsService struct {
	mock.Mock
}

func (m *MockLoadContactsService) ProcessUpload(ctx context.Context, userID int, filename string, payload io.ReadSeeker) error {
	args := m.Called(ctx, userID, filename, payload)
	return args.Error(0)
}

func TestLoadContactsHandler_LoadContactsFile(t *testing.T) {
	nopLogger := zap.NewNop()
	timeout := 500 * time.Millisecond

	tests := []struct {
		name           string
		userID         interface{}
		filename       string
		prepareBody    func(t *testing.T) (*bytes.Buffer, string)
		prepareService func(svc *MockLoadContactsService)
		wantStatus     int
		wantContains   string
	}{
		{
			name:     "no userID",
			userID:   nil,
			filename: "any.csv",
			prepareBody: func(t *testing.T) (*bytes.Buffer, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				_ = w.Close()
				return b, w.FormDataContentType()
			},
			prepareService: func(svc *MockLoadContactsService) {},
			wantStatus:     http.StatusInternalServerError,
			wantContains:   "internal server error",
		},
		{
			name:     "missing file",
			userID:   42,
			filename: "any.csv",
			prepareBody: func(t *testing.T) (*bytes.Buffer, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				// no file part
				_ = w.Close()
				return b, w.FormDataContentType()
			},
			prepareService: func(svc *MockLoadContactsService) {},
			wantStatus:     http.StatusBadRequest,
			wantContains:   "missing file",
		},
		{
			name:     "invalid ext",
			userID:   7,
			filename: "bad.txt",
			prepareBody: func(t *testing.T) (*bytes.Buffer, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				f, err := w.CreateFormFile("file", "bad.txt")
				require.NoError(t, err)
				_, _ = f.Write([]byte("data"))
				_ = w.Close()
				return b, w.FormDataContentType()
			},
			prepareService: func(svc *MockLoadContactsService) {},
			wantStatus:     http.StatusUnprocessableEntity,
			wantContains:   "only csv and xlsx files allowed",
		},
		{
			name:     "invalid content type",
			userID:   8,
			filename: "contacts.csv",
			prepareBody: func(t *testing.T) (*bytes.Buffer, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				f, err := w.CreateFormFile("file", "contacts.csv")
				require.NoError(t, err)
				_, _ = f.Write([]byte("\x00\x01"))
				_ = w.Close()
				return b, w.FormDataContentType()
			},
			prepareService: func(svc *MockLoadContactsService) {},
			wantStatus:     http.StatusUnprocessableEntity,
			wantContains:   "invalid file type",
		},
		{
			name:     "service error",
			userID:   9,
			filename: "foo.csv",
			prepareBody: func(t *testing.T) (*bytes.Buffer, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				f, err := w.CreateFormFile("file", "foo.csv")
				require.NoError(t, err)
				_, _ = f.Write([]byte("a,b,c\n"))
				_ = w.Close()
				return b, w.FormDataContentType()
			},
			prepareService: func(svc *MockLoadContactsService) {
				svc.On("ProcessUpload", mock.Anything, 9, "foo.csv", mock.Anything).
					Return(errors.New("fail")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantContains: "internal server error",
		},
		{
			name:     "success",
			userID:   10,
			filename: "my.csv",
			prepareBody: func(t *testing.T) (*bytes.Buffer, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				f, err := w.CreateFormFile("file", "my.csv")
				require.NoError(t, err)
				_, _ = f.Write([]byte("a,b,c\n"))
				if err != nil {
					return nil, ""
				}
				_ = w.Close()
				return b, w.FormDataContentType()
			},
			prepareService: func(svc *MockLoadContactsService) {
				svc.On("ProcessUpload", mock.Anything, 10, "my.csv", mock.Anything).
					Return(nil).Once()
			},
			wantStatus:   http.StatusAccepted,
			wantContains: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// build handler with mock service
			mockSvc := new(MockLoadContactsService)
			h := handler.NewLoadContactsHandler(mockSvc, nopLogger, timeout)

			// setup mock
			if tc.prepareService != nil {
				tc.prepareService(mockSvc)
			}

			// prepare request body
			body, ct := tc.prepareBody(t)
			req := httptest.NewRequest(http.MethodPost, "/load", body)
			req.Header.Set("Content-Type", ct)
			if tc.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), contextkeys.UserID, tc.userID))
			}

			// execute

			rr := httptest.NewRecorder()
			h.LoadContactsFile(rr, req)
			res := rr.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Fatal("failed to close response.body")
				}
			}(res.Body)

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			if tc.wantContains != "" {
				bts, _ := io.ReadAll(res.Body)
				assert.Contains(t, string(bts), tc.wantContains)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}
