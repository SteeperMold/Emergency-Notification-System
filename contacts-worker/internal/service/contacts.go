package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/phoneutils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/segmentio/kafka-go"
)

// ContactsService processes tasks by downloading files from S3,
// parsing their contents, validating each record, and sending
// them to Kafka. It also deletes the file upon successful processing.
type ContactsService struct {
	kafkaWriter    domain.KafkaWriter
	s3Client       domain.S3Client
	bucket         string
	contextTimeout time.Duration
}

// NewContactsService constructs a new ContactsService.
func NewContactsService(kafkaWriter domain.KafkaWriter, s3Client domain.S3Client, bucket string, timeout time.Duration) *ContactsService {
	return &ContactsService{
		kafkaWriter:    kafkaWriter,
		s3Client:       s3Client,
		bucket:         bucket,
		contextTimeout: timeout,
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
	defer func(Body io.ReadCloser) {
		if cerr := Body.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}(object.Body)

	br := bufio.NewReader(object.Body)

	header, err := br.Peek(512)
	if err != nil && !errors.Is(err, bufio.ErrBufferFull) && !errors.Is(err, io.EOF) {
		return 0, err
	}

	contentType := http.DetectContentType(header)
	switch contentType {
	case "text/plain; charset=utf-8", "text/csv":
		total, err = cs.processCsv(ctx, task.UserID, br)
	case "application/zip":
		total, err = cs.processExcel(ctx, task.UserID, br)
	default:
		return 0, nil
	}

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

var bufPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func (cs *ContactsService) makeMessage(userID int, name string, phone string) ([]byte, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	_, err := fmt.Fprintf(buf, `{"userID":%d,"name":"%s","phone":"%s"}`, userID, name, phone)
	if err != nil {
		return nil, err
	}
	msg := make([]byte, buf.Len())
	copy(msg, buf.Bytes())
	bufPool.Put(buf)
	return msg, nil
}

func (cs *ContactsService) flushBatch(ctx context.Context, batch *[]kafka.Message, total *int32) {
	err := cs.kafkaWriter.WriteMessages(ctx, *batch...)
	if err == nil { // if no error
		atomic.AddInt32(total, int32(len(*batch)))
	}
	*batch = (*batch)[:0]
}

func (cs *ContactsService) validateName(name string) (string, bool) {
	return name, len(name) > 0 && len(name) <= 32
}

func (cs *ContactsService) validatePhone(phone string) (string, bool) {
	formattedNum, err := phoneutils.FormatToE164(phone, phoneutils.RegionRU)
	return formattedNum, err == nil
}
