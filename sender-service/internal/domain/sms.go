package domain

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
