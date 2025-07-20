package domain

type SmsSender interface {
	SendSMS(to, body, notificationID string) error
}

type SendError interface {
	error
	Retryable() bool
}
