package handler_test

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockHealthCheckService struct {
	mock.Mock
}

func (m *MockHealthCheckService) HealthCheck(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockTwilioCallbackService struct {
	mock.Mock
}

func (m *MockTwilioCallbackService) ProcessCallback(ctx context.Context, idStr, status string) error {
	return m.Called(ctx, idStr, status).Error(0)
}
