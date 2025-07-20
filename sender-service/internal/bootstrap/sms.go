package bootstrap

import (
	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/sms"
	"log"
	"time"
)

func NewSmsSender(appEnv string, twilioCfg *TwilioConfig) domain.SmsSender {
	var smsSender domain.SmsSender
	var err error

	switch appEnv {
	case "production":
		smsSender = sms.NewSmsSender(twilioCfg.AccountSID, twilioCfg.AuthToken, twilioCfg.FromNumber, twilioCfg.StatusCallbackEndpoint)
	case "development":
		smsSender, err = sms.NewDevSmsSender("/tmp/sms-dev", twilioCfg.StatusCallbackEndpoint, 0.05, 0.2, 3*time.Second)
	}

	if err != nil {
		log.Fatalf("failed to create sms sender: %v", err)
	}

	return smsSender
}
