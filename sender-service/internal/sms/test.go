package sms

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// TestSendError simulates a structured error from the TestSmsSender.
type TestSendError struct {
	Message   string
	retryable bool
}

func (e TestSendError) Error() string {
	return e.Message
}

func (e TestSendError) Retryable() bool {
	return e.retryable
}

type TestSmsSender struct {
	CallbackBaseURL  string
	FailRate         float64       // 0.0–1.0 chance of send failure
	CallbackDelay    time.Duration // how long until we fire the callback
	CallbackFailRate float64       // 0.0–1.0 chance of callback failure
	rng              *rand.Rand
}

func NewTestSmsSender(callbackBaseURL string, failRate, cbFailRate float64, cbDelay time.Duration) *TestSmsSender {
	return &TestSmsSender{
		CallbackBaseURL:  callbackBaseURL,
		FailRate:         failRate,
		CallbackDelay:    cbDelay,
		CallbackFailRate: cbFailRate,
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (d *TestSmsSender) SendSMS(to, body, notificationID string) error {
	if d.rng.Float64() < d.FailRate {
		return TestSendError{
			Message:   "dev sender: simulated send failure",
			retryable: true,
		}
	}

	callbackFailed := d.rng.Float64() < d.CallbackFailRate

	ts := time.Now().Format("02.01.2006-15:04:05")
	sid := fmt.Sprintf("%s__%s.txt", ts, to)

	go func(sid string) {
		time.Sleep(d.CallbackDelay)

		u, err := url.Parse(d.CallbackBaseURL)
		if err != nil {
			return
		}

		q := u.Query()
		q.Set("notification_id", notificationID)
		u.RawQuery = q.Encode()
		cbURL := u.String()

		var messageStatus string

		if callbackFailed {
			messageStatus = "failed"
		} else {
			messageStatus = "sent"
		}

		form := url.Values{
			"MessageSid":    {sid},
			"MessageStatus": {messageStatus},
			"To":            {to},
			"From":          {"DEV-SENDER"},
		}

		_, _ = http.PostForm(cbURL, form)
	}(sid)

	return nil
}
