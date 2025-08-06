package domain

import "context"

// TwilioCallbackService defines the interface for handling status callbacks from Twilio.
type TwilioCallbackService interface {
	ProcessCallback(ctx context.Context, idStr, status string) error
}

// TwilioRequestValidator defines the interface for validating Twilio's signature.
type TwilioRequestValidator interface {
	Validate(url string, params map[string]string, expectedSignature string) bool
}
