package domain

import "context"

// TwilioCallbackService defines the interface for handling status callbacks from Twilio. .
type TwilioCallbackService interface {
	ProcessCallback(ctx context.Context, idStr, status string) error
}
