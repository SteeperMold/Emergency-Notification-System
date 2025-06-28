package service_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xuri/excelize/v2"
)

type MockS3 struct{ mock.Mock }

func (m *MockS3) GetObjectWithContext(ctx aws.Context, in *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3) PutObjectWithContext(ctx aws.Context, in *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3) DeleteObjectWithContext(ctx aws.Context, in *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error) {
	return m.Called(ctx, in).Get(0).(*s3.DeleteObjectOutput), m.Called(ctx, in).Error(1)
}

type MockKafka struct{ mock.Mock }

func (m *MockKafka) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.Called(ctx, msgs).Error(0)
}

func TestProcessFile_CSV_and_Excel(t *testing.T) {
	bufXLSX := new(bytes.Buffer)
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &[]interface{}{"Carol", "+79123456791"})
	_ = f.SetSheetRow(sheet, "A2", &[]interface{}{"Dave", "+79123456792"})
	_ = f.Write(bufXLSX)
	_ = f.Close()

	tests := []struct {
		name        string
		content     io.Reader
		contentType string
		expectCount int
		s3GetErr    error
	}{
		{
			name:        "valid CSV",
			content:     strings.NewReader("Alice,+79123456789\nBob,+79123456790\n"),
			contentType: "text/csv",
			expectCount: 2,
			s3GetErr:    nil,
		},
		{
			name:        "invalid CSV",
			content:     strings.NewReader("fdsgsfgssa;fsak;ldasdfadskfa.dckm.sasda"),
			contentType: "text/csv",
			expectCount: 0,
			s3GetErr:    nil,
		},
		{
			name:        "invalid s3 get",
			content:     io.NopCloser(nil),
			contentType: "",
			expectCount: 0,
			s3GetErr:    errors.New("s3 fail"),
		},
		{
			name:        "valid XLSX",
			content:     bufXLSX,
			contentType: "application/zip",
			expectCount: 2,
			s3GetErr:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mS3 := new(MockS3)
			mKafka := new(MockKafka)
			svc := service.NewContactsService(mKafka, mS3, "bucket", time.Second)

			task := &domain.Task{S3Key: "key", UserID: 1}
			getOut := &s3.GetObjectOutput{Body: io.NopCloser(tc.content)}

			mS3.On("GetObjectWithContext", mock.Anything, mock.Anything).Return(getOut, tc.s3GetErr)
			if tc.s3GetErr == nil {
				delOut := &s3.DeleteObjectOutput{}
				mS3.On("DeleteObjectWithContext", mock.Anything, mock.Anything).Return(delOut, nil)

				mKafka.On("WriteMessages", mock.Anything, mock.MatchedBy(func(msgs []kafka.Message) bool {
					return len(msgs) > 0
				})).Return(nil)
			}

			total, err := svc.ProcessFile(context.Background(), task)

			if tc.s3GetErr != nil {
				assert.Error(t, err)
				assert.Zero(t, total)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectCount, total)
			}
		})
	}
}
