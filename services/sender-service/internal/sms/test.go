package sms

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

// TestSendError simulates a structured error from the TestSmsSender.
type TestSendError struct {
	Message   string
	retryable bool
}

// Error returns the error message for TestSendError.
func (e TestSendError) Error() string {
	return e.Message
}

// Retryable indicates whether the error is retryable or not.
func (e TestSendError) Retryable() bool {
	return e.retryable
}

// TestSmsSender simulates an SMS provider for testing purposes.
// It can be configured to randomly fail sending messages or callbacks.
// TestSmsSender doesn't write messages anywhere, it only sends callback with sms status.
type TestSmsSender struct {
	CallbackBaseURL  string
	FailRate         float64       // 0.0–1.0 chance of send failure
	CallbackDelay    time.Duration // how long until we fire the callback
	CallbackFailRate float64       // 0.0–1.0 chance of callback failure
	rng              *rand.Rand
	jobs             chan callbackJob
}

type callbackJob struct {
	to             string
	notificationID string
	sid            string
	callbackFailed bool
}

// NewTestSmsSender returns a new TestSmsSender with the given parameters.
func NewTestSmsSender(callbackBaseURL string, failRate, cbFailRate float64, cbDelay time.Duration) *TestSmsSender {
	s := &TestSmsSender{
		CallbackBaseURL:  callbackBaseURL,
		FailRate:         failRate,
		CallbackDelay:    cbDelay,
		CallbackFailRate: cbFailRate,
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
		jobs:             make(chan callbackJob, 1000),
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go s.worker()
	}

	return s
}

// SendSMS simulates sending an SMS message.
// It may return a retryable error depending on the configured FailRate.
// If the sending is successful, a delivery callback will eventually be sent.
func (d *TestSmsSender) SendSMS(to, body, notificationID string) error {
	if d.rng.Float64() < d.FailRate {
		return TestSendError{"dev sender: simulated send failure", true}
	}

	sid := fmt.Sprintf("%s__%s.txt", time.Now().Format("02.01.2006-15:04:05"), to)
	cbFailed := d.rng.Float64() < d.CallbackFailRate

	d.jobs <- callbackJob{to, notificationID, sid, cbFailed}
	return nil
}

func (d *TestSmsSender) worker() {
	for job := range d.jobs {
		time.Sleep(d.CallbackDelay)

		u, err := url.Parse(d.CallbackBaseURL)
		if err != nil {
			continue
		}

		q := u.Query()
		q.Set("notification_id", job.notificationID)
		u.RawQuery = q.Encode()
		cbURL := u.String()

		status := "sent"
		if job.callbackFailed {
			status = "failed"
		}

		form := url.Values{
			"MessageSid":    {job.sid},
			"MessageStatus": {status},
			"To":            {job.to},
			"From":          {"DEV-SENDER"},
		}

		_, _ = http.PostForm(cbURL, form)
	}
}
