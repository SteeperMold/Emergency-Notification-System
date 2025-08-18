package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBasicNotificationFlow(t *testing.T) {
	initialCount := FetchSentNotificationsCount(t)

	accessToken := SendSignupRequest(t, "test@e2e.com", "123456789admin")
	templateID := PostTemplateRequest(t, accessToken, "Test template name", "Test template body")
	PostContactRequest(t, accessToken, "Test contact name", "+79123456789")
	SendNotificationRequest(t, accessToken, templateID)

	deadline := time.Now().Add(30 * time.Second)
	ntfCreated := false

	for time.Now().Before(deadline) {
		count := FetchSentNotificationsCount(t)

		if count-initialCount == 1 {
			ntfCreated = true
			break
		}

		time.Sleep(1 * time.Second)
	}

	require.True(t, ntfCreated)
}
