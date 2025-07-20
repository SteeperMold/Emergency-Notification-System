package domain

import "context"

type TwilioCallbackService interface {
	ProcessCallback(ctx context.Context, idStr, status string) error
}
