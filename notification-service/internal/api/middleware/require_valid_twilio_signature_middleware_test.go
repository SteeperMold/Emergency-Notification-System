package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRequestValidator struct {
	mock.Mock
}

func (m *MockRequestValidator) Validate(url string, params map[string]string, expectedSignature string) bool {
	return m.Called(url, params, expectedSignature).Bool(0)
}

func TestRequireValidTwilioSignatureMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		form              url.Values
		signatureHeader   string
		mockValid         bool
		expectStatus      int
		expectNextInvoked bool
	}{
		{
			name:              "valid signature passes",
			form:              url.Values{"MessageSid": {"abc"}, "MessageStatus": {"delivered"}},
			signatureHeader:   "valid-signature",
			mockValid:         true,
			expectStatus:      http.StatusOK,
			expectNextInvoked: true,
		},
		{
			name:              "invalid signature rejected",
			form:              url.Values{"MessageSid": {"abc"}, "MessageStatus": {"delivered"}},
			signatureHeader:   "invalid-signature",
			mockValid:         false,
			expectStatus:      http.StatusForbidden,
			expectNextInvoked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := new(MockRequestValidator)
			baseURL := "https://example.com"
			path := "/twilio/callback"
			fullURL := baseURL + path

			var body *strings.Reader
			if tt.form != nil {
				body = strings.NewReader(tt.form.Encode())
			} else {
				body = nil
			}

			req := httptest.NewRequest(http.MethodPost, path, body)
			req.Header.Set("X-Twilio-Signature", tt.signatureHeader)
			if tt.form != nil {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			if tt.form != nil {
				paramMap := map[string]string{}
				for k, v := range tt.form {
					paramMap[k] = v[0]
				}
				mockValidator.On("Validate", fullURL, paramMap, tt.signatureHeader).Return(tt.mockValid).Once()
			}

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			mw := middleware.RequireValidTwilioSignatureMiddleware(baseURL, mockValidator)
			handler := mw(nextHandler)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectStatus, rr.Code)
			assert.Equal(t, tt.expectNextInvoked, nextCalled)
			mockValidator.AssertExpectations(t)
		})
	}
}
