package domain

import openapi "github.com/twilio/twilio-go/rest/api/v2010"

// SmsSender defines an interface for sending SMS messages.
type SmsSender interface {
	SendSMS(to, body, notificationID string) error
}

// SendError represents an error returned from an SMS sending operation
// that can indicate whether the error is retryable.
type SendError interface {
	error
	Retryable() bool
}

// TwilioAPI defines the minimal interface for interacting with Twilio's API.
type TwilioAPI interface {
	CreateMessage(params *openapi.CreateMessageParams) (*openapi.ApiV2010Message, error)
}
