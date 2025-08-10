package service_test

import (
	"context"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/notification-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHealthCheckService_HealthCheck(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(db *MockDBConn, kf *MockKafkaFactory)
		wantErr   error
	}{
		{
			name: "all healthy",
			mockSetup: func(db *MockDBConn, kf *MockKafkaFactory) {
				db.
					On("Ping", mock.Anything).
					Return(nil).
					Once()
				kf.
					On("Ping", mock.Anything).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "db ping fails",
			mockSetup: func(db *MockDBConn, kf *MockKafkaFactory) {
				db.
					On("Ping", mock.Anything).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "kafka ping fails",
			mockSetup: func(db *MockDBConn, kf *MockKafkaFactory) {
				db.
					On("Ping", mock.Anything).
					Return(nil).
					Once()
				kf.
					On("Ping", mock.Anything).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dbMock := new(MockDBConn)
			kfMock := new(MockKafkaFactory)
			tc.mockSetup(dbMock, kfMock)

			svc := service.NewHealthCheckService(dbMock, kfMock)
			err := svc.HealthCheck(context.Background())

			if tc.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			dbMock.AssertExpectations(t)
			kfMock.AssertExpectations(t)
		})
	}
}
