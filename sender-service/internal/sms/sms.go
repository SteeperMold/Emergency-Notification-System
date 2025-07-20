package sms

import (
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"net/http"
	"net/url"
)

type TwilioSendError struct {
	Code      int
	Message   string
	retryable bool
}

func (e TwilioSendError) Error() string {
	return e.Message
}

func (e TwilioSendError) Retryable() bool {
	return e.retryable
}

type SmsSender struct {
	client          *twilio.RestClient
	fromNumber      string
	callbackBaseURL string
}

func NewSmsSender(accountSID, authToken, fromNumber, callbackBaseURL string) *SmsSender {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	return &SmsSender{
		client:          client,
		fromNumber:      fromNumber,
		callbackBaseURL: callbackBaseURL,
	}
}

func (s *SmsSender) SendSMS(to, body, notificationID string) error {
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
