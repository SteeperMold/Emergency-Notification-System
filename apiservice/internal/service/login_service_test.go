package service_test

import (
	"context"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginService_GetUserByEmail(t *testing.T) {
	user := &models.User{ID: 123, Email: "test@test.com"}

	m := new(MockUserRepository)
	m.
		On("GetUserByEmail", mock.Anything, "test@test.com").
		Return(user, nil).
		Once()
	svc := service.NewLoginService(m)

	res, err := svc.GetUserByEmail(context.Background(), "test@test.com")
	assert.NoError(t, err)
	assert.Equal(t, user, res)
	m.AssertExpectations(t)
}

func TestLoginService_CompareCredentials(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user := &models.User{PasswordHash: string(hash)}
	svc := service.NewLoginService(nil)

	t.Run("correct password", func(t *testing.T) {
		request := &domain.LoginRequest{Password: "secret"}
		assert.True(t, svc.CompareCredentials(user, request))
	})

	t.Run("incorrect password", func(t *testing.T) {
		request := &domain.LoginRequest{Password: "wrong"}
		assert.False(t, svc.CompareCredentials(user, request))
	})
}
