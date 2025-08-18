package sms

import (
	"net/http"
	"net/url"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioSendError represents an error returned by the Twilio SMS sender.
// It includes an error code, message, and a flag indicating if the error is retryable.
type TwilioSendError struct {
	Code      int
	Message   string
	retryable bool
}

// Error returns the error message for TwilioSendError.
func (e TwilioSendError) Error() string {
	return e.Message
}

// Retryable indicates whether the TwilioSendError is considered retryable.
func (e TwilioSendError) Retryable() bool {
	return e.retryable
}

// Sender sends SMS messages using Twilio.
// It also registers a status callback for delivery reporting.
type Sender struct {
	client          *twilio.RestClient
	fromNumber      string
	callbackBaseURL string
}

// NewSmsSender initializes and returns a new SmsSender.
func NewSmsSender(accountSID, authToken, fromNumber, callbackBaseURL string) *Sender {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	return &Sender{
		client:          client,
		fromNumber:      fromNumber,
		callbackBaseURL: callbackBaseURL,
	}
}

// SendSMS sends an SMS message using Twilio's API.
// It sets a status callback for delivery tracking and returns a TwilioSendError
// if sending fails or if Twilio returns an error code.
func (s *Sender) SendSMS(to, body, notificationID string) error {
	cb, err := url.Parse(s.callbackBaseURL)
	if err != nil {
		return err
	}
	q := cb.Query()
	q.Set("notification_id", notificationID)
	cb.RawQuery = q.Encode()
	cbURL := cb.String()

	params := &api.CreateMessageParams{}
	params.SetFrom(s.fromNumber)
	params.SetStatusCallback(cbURL)
	params.SetTo(to)
	params.SetBody(body)

	resp, err := s.client.Api.CreateMessage(params)
	if err != nil {
		// assume low-level errors (e.g. network) are retryable
		return TwilioSendError{
			Code:      http.StatusServiceUnavailable,
			Message:   err.Error(),
			retryable: true,
		}
	}
	if resp.ErrorCode != nil {
		retryable := isRetryableTwilioError(*resp.ErrorCode)
		return TwilioSendError{
			Code:      *resp.ErrorCode,
			Message:   *resp.ErrorMessage,
			retryable: retryable,
		}
	}

	return nil
}

func isRetryableTwilioError(code int) bool {
	switch code {
	case 20429, // Too many requests
		30001, // Queue overflow
		30002, // Account suspended
		30006, // Landline or unreachable
		30008: // Unknown error
		return true
	}
	return false
}
