package service

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/mock"
)

type MockContactsRepository struct {
	mock.Mock
}

func (m *MockContactsRepository) SaveContacts(ctx context.Context, contacts []*models.Contact) error {
	return m.Called(ctx, contacts).Error(0)
}

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3Client) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3Client) DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}
