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
