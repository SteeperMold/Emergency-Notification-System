package service

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/contacts-worker/internal/domain"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestContactsService_ProcessFile(t *testing.T) {
	tests := []struct {
		name          string
		body          []byte
		setupMocks    func(repo *MockContactsRepository, s3c *MockS3Client)
		expectedTotal int
		expectedErr   bool
	}{
		{
			name: "S3 GetObject error",
			body: nil,
			setupMocks: func(repo *MockContactsRepository, s3c *MockS3Client) {
				s3c.
					On("GetObjectWithContext", mock.Anything, mock.MatchedBy(func(input *s3.GetObjectInput) bool {
						return aws.StringValue(input.Bucket) == "test-bucket" && aws.StringValue(input.Key) == "key.csv"
					}), mock.Anything).
					Return((*s3.GetObjectOutput)(nil), assert.AnError).
					Once()
			},
			expectedTotal: 0,
			expectedErr:   true,
		},
		{
			name: "Unknown content type",
			body: []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}, // GIF89a magic
			setupMocks: func(repo *MockContactsRepository, s3c *MockS3Client) {
				s3c.
					On("GetObjectWithContext", mock.Anything, mock.Anything, mock.Anything).
					Return(&s3.GetObjectOutput{
						Body: io.NopCloser(bytes.NewReader([]byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61})),
					}, nil).
					Once()
			},
			expectedTotal: 0,
			expectedErr:   false,
		},
		{
			name: "CSV processing success",
			body: []byte("name,email\nAlice,+79123456789\nBob,+79123456788\n"),
			setupMocks: func(repo *MockContactsRepository, s3c *MockS3Client) {
				// expect save of two contacts
				// we’ll get two batches, one per worker that got a row
				repo.
					On("SaveContacts", mock.Anything, mock.Anything).
					Return(nil).
					Twice()
				s3c.
					On("GetObjectWithContext", mock.Anything, mock.Anything, mock.Anything).
					Return(&s3.GetObjectOutput{
						Body: io.NopCloser(bytes.NewReader([]byte("name,email\nAlice,+79123456789\nBob,+79123456788\n"))),
					}, nil).
					Once()
				// S3 delete
				s3c.
					On("DeleteObjectWithContext", mock.Anything, mock.MatchedBy(func(input *s3.DeleteObjectInput) bool {
						return aws.StringValue(input.Bucket) == "test-bucket" && aws.StringValue(input.Key) == "key.csv"
					}), mock.Anything).
					Return(&s3.DeleteObjectOutput{}, nil).
					Once()
			},
			expectedTotal: 2,
			expectedErr:   false,
		},
		{
			name: "Repository save error",
			body: []byte("name,email\nAlice,+79123456789\nBob,+79123456788\n"),
			setupMocks: func(repo *MockContactsRepository, s3c *MockS3Client) {
				repo.
					On("SaveContacts", mock.Anything, mock.Anything).
					Return(assert.AnError).
					Once()
				s3c.
					On("GetObjectWithContext", mock.Anything, mock.Anything, mock.Anything).
					Return(&s3.GetObjectOutput{
						Body: io.NopCloser(bytes.NewReader([]byte("name,email\nAlice,+79123456789\nBob,+79123456788\n"))),
					}, nil).
					Once()
			},
			expectedTotal: 0,
			expectedErr:   true,
		},
		{
			name: "Delete error",
			body: []byte("name,email\nAlice,+79123456789\nBob,+79123456788\n"),
			setupMocks: func(repo *MockContactsRepository, s3c *MockS3Client) {
				repo.
					On("SaveContacts", mock.Anything, mock.Anything).
					Return(nil)
				s3c.
					On("GetObjectWithContext", mock.Anything, mock.Anything, mock.Anything).
					Return(&s3.GetObjectOutput{
						Body: io.NopCloser(bytes.NewReader([]byte("name,email\nAlice,+79123456789\nBob,+79123456788\n"))),
					}, nil).
					Once()
				s3c.
					On("DeleteObjectWithContext", mock.Anything, mock.Anything, mock.Anything).
					Return((*s3.DeleteObjectOutput)(nil), assert.AnError).
					Once()
			},
			expectedTotal: 2,
			expectedErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			repo := new(MockContactsRepository)
			s3c := new(MockS3Client)
			bucket := "test-bucket"
			cs := NewContactsService(repo, s3c, bucket, 5*time.Second, 100)

			tt.setupMocks(repo, s3c)

			total, err := cs.ProcessFile(context.Background(), &domain.Task{UserID: 123, S3Key: "key.csv"})

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedTotal, total)
			repo.AssertExpectations(t)
			s3c.AssertExpectations(t)
		})
	}
}
