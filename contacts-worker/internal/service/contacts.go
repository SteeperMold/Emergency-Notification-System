package service

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/domain"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ContactsService processes tasks by downloading files from S3,
// parsing their contents, validating each record, and sending
// them to Kafka. It also deletes the file upon successful processing.
type ContactsService struct {
	repository     domain.ContactsRepository
	s3Client       domain.S3Client
	bucket         string
	contextTimeout time.Duration
	batchSize      int
}

// NewContactsService constructs a new ContactsService.
func NewContactsService(r domain.ContactsRepository, s3Client domain.S3Client, bucket string, timeout time.Duration, batchSize int) *ContactsService {
	return &ContactsService{
		repository:     r,
		s3Client:       s3Client,
		bucket:         bucket,
		contextTimeout: timeout,
		batchSize:      batchSize,
	}
}

// ProcessFile retrieves the file specified by task.S3Key from S3, determines
// its type by magic bytes, and processes CSV or Excel accordingly. It returns
// the number of successfully processed records and any error encountered.
func (cs *ContactsService) ProcessFile(ctx context.Context, task *domain.Task) (total int, err error) {
	ctx, cancel := context.WithTimeout(ctx, cs.contextTimeout)
	defer cancel()

	object, err := cs.getFileFromS3(ctx, task.S3Key)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cerr := object.Body.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	br := bufio.NewReader(object.Body)

	header, err := br.Peek(512)
	if err != nil && !errors.Is(err, bufio.ErrBufferFull) && !errors.Is(err, io.EOF) {
		return 0, err
	}

	var rowProvider RowProvider
	switch http.DetectContentType(header) {
	case "text/plain; charset=utf-8", "text/csv":
		rowProvider, err = cs.createCsvRowProvider(br)
	case "application/zip":
		rowProvider, err = cs.createExcelRowProvider(br)
	default:
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	total, err = cs.ingestAndSave(ctx, task.UserID, rowProvider)
	if err != nil {
		return 0, err
	}

	_, err = cs.deleteFileFromS3(ctx, task.S3Key)
	if err != nil {
		return total, err
	}

	return total, nil
}

func (cs *ContactsService) getFileFromS3(ctx context.Context, s3key string) (*s3.GetObjectOutput, error) {
	return cs.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(cs.bucket),
		Key:    aws.String(s3key),
	})
}

func (cs *ContactsService) deleteFileFromS3(ctx context.Context, s3Key string) (*s3.DeleteObjectOutput, error) {
	return cs.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(cs.bucket),
		Key:    aws.String(s3Key),
	})
}
