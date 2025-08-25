package sms

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

func newTestSender(api *MockTwilioAPI) *Sender {
	return &Sender{
		twilioAPI:       api,
		fromNumber:      "+199999",
		callbackBaseURL: "http://callback.local/status",
	}
}

func TestSendSMS(t *testing.T) {
	tests := map[string]struct {
		apiResp    *api.ApiV2010Message
		apiErr     error
		expectErr  bool
		expectCode int
		retryable  bool
	}{
		"success": {
			apiResp: &api.ApiV2010Message{Sid: toStrPtr("SM123")},
			apiErr:  nil,
		},
		"network error": {
			apiResp:    nil,
			apiErr:     assert.AnError,
			expectErr:  true,
			expectCode: http.StatusServiceUnavailable,
			retryable:  true,
		},
		"twilio non-retryable error": {
			apiResp: &api.ApiV2010Message{
				ErrorCode:    toIntPtr(21614), // "to" number not valid
				ErrorMessage: toStrPtr("Invalid To number"),
			},
			apiErr:     nil,
			expectErr:  true,
			expectCode: 21614,
			retryable:  false,
		},
		"twilio retryable error": {
			apiResp: &api.ApiV2010Message{
				ErrorCode:    toIntPtr(20429), // Too many requests
				ErrorMessage: toStrPtr("Rate limit exceeded"),
			},
			apiErr:     nil,
			expectErr:  true,
			expectCode: 20429,
			retryable:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			apiMock := &MockTwilioAPI{}
			apiMock.
				On("CreateMessage", mock.Anything).
				Return(tc.apiResp, tc.apiErr).
				Once()

			s := newTestSender(apiMock)

			err := s.SendSMS("+100000", "hello", "notif-123")

			if tc.expectErr {
				assert.Error(t, err)
				var twErr TwilioSendError
				ok := errors.As(err, &twErr)
				if ok {
					assert.Equal(t, tc.expectCode, twErr.Code)
					assert.Equal(t, tc.retryable, twErr.Retryable())
				}
			} else {
				assert.NoError(t, err)
			}

			apiMock.AssertExpectations(t)
		})
	}
}

func toIntPtr(v int) *int {
	return &v
}

func toStrPtr(s string) *string {
	return &s
}
