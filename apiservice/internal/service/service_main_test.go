package service_test

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	u := args.Get(0)
	if u != nil {
		return u.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	u := args.Get(0)
	if u != nil {
		return u.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if user := args.Get(0); user != nil {
		return user.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) PutObjectWithContext(ctx context.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input)
	out, _ := args.Get(0).(*s3.PutObjectOutput)
	return out, args.Error(1)
}

type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}
