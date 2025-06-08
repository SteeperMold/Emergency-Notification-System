package service_test

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
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
