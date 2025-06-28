// consumers_test.go
package consumers_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/adapter/consumers"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/domain"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockService struct{ mock.Mock }

func (m *MockService) ProcessFile(ctx context.Context, task *domain.Task) (int, error) {
	args := m.Called(ctx, task)
	return args.Int(0), args.Error(1)
}

type MockReader struct{ mock.Mock }

func (m *MockReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func TestStartConsumer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	validTask := domain.Task{UserID: 1, S3Key: "key"}
	validBytes, _ := json.Marshal(validTask)
	invalidBytes := []byte("{bad json}")
	msgValid := kafka.Message{Value: validBytes}
	msgInvalid := kafka.Message{Value: invalidBytes}

	cases := []struct {
		name          string
		sequenceFetch []struct {
			msg kafka.Message
			err error
		}
		processResult []struct {
			count int
			err   error
		}
		commitErr      error
		wantErr        bool
		wantProcessLen int
	}{
		{
			name: "success then fetch error",
			sequenceFetch: []struct {
				msg kafka.Message
				err error
			}{
				{msgValid, nil},
				{kafka.Message{}, errors.New("done")},
			},
			processResult: []struct {
				count int
				err   error
			}{{2, nil}},
			commitErr:      nil,
			wantErr:        true,
			wantProcessLen: 1,
		},
		{
			name: "invalid JSON then valid",
			sequenceFetch: []struct {
				msg kafka.Message
				err error
			}{
				{msgInvalid, nil},
				{msgValid, nil},
				{kafka.Message{}, errors.New("stop")},
			},
			processResult: []struct {
				count int
				err   error
			}{{3, nil}},
			commitErr:      nil,
			wantErr:        true,
			wantProcessLen: 1,
		},
		{
			name: "process error but commit OK",
			sequenceFetch: []struct {
				msg kafka.Message
				err error
			}{
				{msgValid, nil},
				{kafka.Message{}, errors.New("stop")},
			},
			processResult: []struct {
				count int
				err   error
			}{{0, errors.New("fail")}},
			commitErr:      nil,
			wantErr:        true,
			wantProcessLen: 1,
		},
		{
			name: "commit error",
			sequenceFetch: []struct {
				msg kafka.Message
				err error
			}{
				{msgValid, nil},
			},
			processResult: []struct {
				count int
				err   error
			}{{1, nil}},
			commitErr:      errors.New("commit fail"),
			wantErr:        true,
			wantProcessLen: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := new(MockService)
			rd := new(MockReader)
			logger := zap.NewNop()
			cc := consumers.NewContactsConsumer(svc, rd, logger)

			for _, sf := range tc.sequenceFetch {
				rd.On("FetchMessage", mock.Anything).
					Return(sf.msg, sf.err).Once()
				if sf.err == nil {
					rd.On("CommitMessages", mock.Anything, []kafka.Message{sf.msg}).
						Return(tc.commitErr).Once()
					if sf.msg.Value != nil {
						svc.
							On("ProcessFile", mock.Anything, mock.Anything).
							Return(tc.processResult[0].count, tc.processResult[0].err).
							Once()
					}
				}
			}

			err := cc.StartConsumer(ctx)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
