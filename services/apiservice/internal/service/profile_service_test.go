package service_test

import (
	"context"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProfileService_GetUserByID(t *testing.T) {
	user := &models.User{ID: 123, Email: "test@test.com"}

	m := new(MockUserRepository)
	m.
		On("GetUserByID", mock.Anything, 123).
		Return(user, nil).
		Once()
	svc := service.NewProfileService(m)

	res, err := svc.GetUserByID(context.Background(), 123)
	assert.NoError(t, err)
	assert.Equal(t, user, res)
	m.AssertExpectations(t)
}
