package service_test

import (
	"context"
	"testing"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/service"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendNotificationService_SendNotification(t *testing.T) {
	userID := 42
	tmplID := 123

	tmpl := &models.Template{ID: tmplID, UserID: userID, Body: "Hello {{.Name}}"}
	contacts := []*models.Contact{
		{ID: 1, UserID: userID, Name: "A", Phone: "+100"},
		{ID: 2, UserID: userID, Name: "B", Phone: "+200"},
		{ID: 3, UserID: userID, Name: "C", Phone: "+300"},
	}

	tests := []struct {
		name                 string
		contactsPerMsg       int
		setupMocks           func(cr *MockContactsRepository, tr *MockTemplateRepository, kw *MockKafkaWriter)
		wantErr              error
		expectedKafkaBatches int
	}{
		{
			name:           "template error",
			contactsPerMsg: 2,
			setupMocks: func(cr *MockContactsRepository, tr *MockTemplateRepository, kw *MockKafkaWriter) {
				tr.
					On("GetTemplateByID", mock.Anything, userID, tmplID).
					Return((*models.Template)(nil), assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name:           "contacts error",
			contactsPerMsg: 2,
			setupMocks: func(cr *MockContactsRepository, tr *MockTemplateRepository, kw *MockKafkaWriter) {
				tr.
					On("GetTemplateByID", mock.Anything, userID, tmplID).
					Return(tmpl, nil).
					Once()
				cr.
					On("GetContactsByUserID", mock.Anything, userID).
					Return(([]*models.Contact)(nil), assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name:           "no contacts",
			contactsPerMsg: 2,
			setupMocks: func(cr *MockContactsRepository, tr *MockTemplateRepository, kw *MockKafkaWriter) {
				tr.
					On("GetTemplateByID", mock.Anything, userID, tmplID).
					Return(tmpl, nil).
					Once()
				cr.
					On("GetContactsByUserID", mock.Anything, userID).
					Return(([]*models.Contact)(nil), nil).
					Once()
			},
			wantErr: domain.ErrContactNotExists,
		},
		{
			name:           "kafka write failure",
			contactsPerMsg: 2,
			setupMocks: func(cr *MockContactsRepository, tr *MockTemplateRepository, kw *MockKafkaWriter) {
				tr.
					On("GetTemplateByID", mock.Anything, userID, tmplID).
					Return(tmpl, nil).
					Once()
				cr.
					On("GetContactsByUserID", mock.Anything, userID).
					Return(contacts, nil).
					Once()
				// first batch fails
				kw.
					On("WriteMessages", mock.Anything, mock.MatchedBy(func(msgs []kafka.Message) bool {
						return len(msgs) == 1
					})).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name:           "successful chunking",
			contactsPerMsg: 2,
			setupMocks: func(cr *MockContactsRepository, tr *MockTemplateRepository, kw *MockKafkaWriter) {
				tr.
					On("GetTemplateByID", mock.Anything, userID, tmplID).
					Return(tmpl, nil).
					Once()
				cr.
					On("GetContactsByUserID", mock.Anything, userID).
					Return(contacts, nil).
					Once()

				// Expect ceil(3/2)=2 calls to WriteMessages
				kw.
					On("WriteMessages", mock.Anything, mock.Anything).
					Return(nil).
					Twice()
			},
			wantErr:              nil,
			expectedKafkaBatches: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cr := new(MockContactsRepository)
			tr := new(MockTemplateRepository)
			kw := new(MockKafkaWriter)
			tc.setupMocks(cr, tr, kw)

			svc := service.NewSendNotificationService(cr, tr, kw, tc.contactsPerMsg)
			err := svc.SendNotification(context.Background(), userID, tmplID)

			if tc.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedKafkaBatches > 0 {
				kw.AssertNumberOfCalls(t, "WriteMessages", tc.expectedKafkaBatches)
			}

			cr.AssertExpectations(t)
			tr.AssertExpectations(t)
			kw.AssertExpectations(t)
		})
	}
}
