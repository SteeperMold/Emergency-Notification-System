package sms

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// DevSendError simulates a structured error from the DevSmsSender.
type DevSendError struct {
	Message   string
	retryable bool
}

func (e DevSendError) Error() string {
	return e.Message
}

func (e DevSendError) Retryable() bool {
	return e.retryable
}

type DevSmsSender struct {
	Dir              string
	CallbackBaseURL  string
	FailRate         float64       // 0.0–1.0 chance of send failure
	CallbackDelay    time.Duration // how long until we fire the callback
	CallbackFailRate float64       // 0.0–1.0 chance of callback failure
	rng              *rand.Rand
}

func NewDevSmsSender(dir, callbackBaseURL string, failRate, cbFailRate float64, cbDelay time.Duration) (*DevSmsSender, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &DevSmsSender{
		Dir:              dir,
		CallbackBaseURL:  callbackBaseURL,
		FailRate:         failRate,
		CallbackDelay:    cbDelay,
		CallbackFailRate: cbFailRate,
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (d *DevSmsSender) SendSMS(to, body, notificationID string) error {
	if d.rng.Float64() < d.FailRate {
		return DevSendError{
			Message:   "dev sender: simulated send failure",
			retryable: true,
		}
	}

	callbackFailed := d.rng.Float64() < d.CallbackFailRate

	ts := time.Now().Format("02.01.2006-15:04:05")
	filename := fmt.Sprintf("%s__%s.txt", ts, to)

	if !callbackFailed {
		path := filepath.Join(d.Dir, filename)
		contents := fmt.Sprintf("To: %s\n\n%s\n\n%s", to, body, ts)
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			return DevSendError{
				Message:   fmt.Sprintf("failed to write sms file: %v", err),
				retryable: false,
			}
		}
	}

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
	}(filename)

	return nil
}
