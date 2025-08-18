package bootstrap

import (
	"log"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/services/sender-service/internal/sms"
)

// NewSmsSender initializes and returns a domain.SmsSender implementation
// based on the provided application environment.
//
// In "production", it returns a real Twilio-backed sender.
// In "development", it returns a file-based sender that simulates SMS delivery and saves messages locally.
// In "test", it returns an in-memory sender that simulates sending and callbacks.
func NewSmsSender(appEnv string, twilioCfg *TwilioConfig) domain.SmsSender {
	var smsSender domain.SmsSender
	var err error

	switch appEnv {
	case "production":
		smsSender = sms.NewSmsSender(twilioCfg.AccountSID, twilioCfg.AuthToken, twilioCfg.FromNumber, twilioCfg.StatusCallbackEndpoint)
	case "development":
		smsSender, err = sms.NewDevSmsSender("/tmp/sms-dev", twilioCfg.StatusCallbackEndpoint, 0.05, 0.2, 3*time.Second)
	case "test":
		smsSender = sms.NewTestSmsSender(twilioCfg.StatusCallbackEndpoint, 0.05, 0.2, 3*time.Second)
	}

	if err != nil {
		log.Fatalf("failed to create sms sender: %v", err)
	}

	return smsSender
}
