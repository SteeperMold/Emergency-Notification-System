package sms

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDevSmsSender_SendSMS(t *testing.T) {
	dir, err := os.MkdirTemp("", "dev-sms-test")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	tests := map[string]struct {
		failRate         float64
		callbackFailRate float64
		sourceValues     []int64
		expectErr        bool
		expectFile       bool
	}{
		"success": {
			failRate:         0,
			callbackFailRate: 0,
			sourceValues:     []int64{rand.Int63()},
			expectErr:        false,
			expectFile:       true,
		},
		"send fail": {
			failRate:         1,
			callbackFailRate: 0,
			sourceValues:     []int64{0},
			expectErr:        true,
			expectFile:       false,
		},
		"callback fail": {
			failRate:         0,
			callbackFailRate: 1,
			sourceValues:     []int64{rand.Int63(), 0},
			expectErr:        false,
			expectFile:       false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sender, err := NewDevSmsSender(dir, "http://localhost", tc.failRate, tc.callbackFailRate, 10*time.Millisecond)
			assert.NoError(t, err)

			sender.rng = rand.New(&fixedSource{values: tc.sourceValues})

			to := "+123456789"
			body := "Hello dev"
			notifID := "notif-1"

			err = sender.SendSMS(to, body, notifID)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			files, _ := filepath.Glob(filepath.Join(dir, "*__"+to+".txt"))
			if tc.expectFile {
				assert.Len(t, files, 1)
				content, _ := os.ReadFile(files[0])
				assert.Contains(t, string(content), to)
				assert.Contains(t, string(content), body)
			} else {
				assert.Len(t, files, 0)
			}

			err = os.RemoveAll(dir)
			assert.NoError(t, err)
			err = os.MkdirAll(dir, 0o755)
			assert.NoError(t, err)
		})
	}
}
