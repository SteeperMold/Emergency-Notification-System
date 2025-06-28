package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoadContactsService_ProcessUpload(t *testing.T) {
	bucket := "test-bucket"
	filename := "contacts.csv"
	userID := 123

	tests := []struct {
		name      string
		mockSetup func(s3 *MockS3Client, kafka *MockKafkaWriter, capturedKey *string)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(s3c *MockS3Client, kw *MockKafkaWriter, capturedKey *string) {
				s3c.On("PutObjectWithContext", mock.Anything, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
					key := aws.StringValue(input.Key)
					// should have prefix "contacts/" and suffix filename
					if !strings.HasPrefix(key, "contacts/") || !strings.HasSuffix(key, "_"+filename) {
						return false
					}
					*capturedKey = key
					// Body should be the payload reader - we can read first few bytes
					buf := make([]byte, 5)
					_, err := input.Body.Read(buf)
					return err == nil
				})).Return(&s3.PutObjectOutput{}, nil).Once()

				kw.On("WriteMessages", mock.Anything, mock.MatchedBy(func(msgs []kafka.Message) bool {
					if len(msgs) != 1 {
						return false
					}
					// Unmarshal JSON
					var task domain.LoadContactsTask
					err := json.Unmarshal(msgs[0].Value, &task)
					return err == nil && task.UserID == userID && task.S3Key == *capturedKey
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "s3 error",
			mockSetup: func(s3c *MockS3Client, kw *MockKafkaWriter, _ *string) {
				s3c.On("PutObjectWithContext", mock.Anything, mock.Anything).
					Return(nil, errors.New("s3 failure")).Once()
				// kafka writer should not be called
			},
			wantErr: errors.New("s3 failure"),
		},
		{
			name: "kafka error",
			mockSetup: func(s3c *MockS3Client, kw *MockKafkaWriter, capturedKey *string) {
				s3c.On("PutObjectWithContext", mock.Anything, mock.Anything).
					Return(&s3.PutObjectOutput{}, nil).Once()
				*capturedKey = "dummy"
				kw.On("WriteMessages", mock.Anything, mock.Anything).
					Return(errors.New("kafka failure")).Once()
			},
			wantErr: errors.New("kafka failure"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s3Mock := new(MockS3Client)
			kafkaMock := new(MockKafkaWriter)
			var capturedKey string
			tc.mockSetup(s3Mock, kafkaMock, &capturedKey)

			svc := service.NewLoadContactsService(s3Mock, bucket, kafkaMock)
			// provide a simple payload
			payload := strings.NewReader("data")
			err := svc.ProcessUpload(context.Background(), userID, filename, payload)
			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			s3Mock.AssertExpectations(t)
			kafkaMock.AssertExpectations(t)
		})
	}
}
