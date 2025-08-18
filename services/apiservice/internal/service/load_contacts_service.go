package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/services/apiservice/internal/domain"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/segmentio/kafka-go"
)

// LoadContactsService uploads contact files to S3 and enqueues a processing task.
type LoadContactsService struct {
	s3Client    domain.S3Client
	bucket      string
	kafkaWriter domain.KafkaWriter
}

// NewLoadContactsService constructs a LoadContactsService.
func NewLoadContactsService(s3Client domain.S3Client, bucket string, kafkaWriter domain.KafkaWriter) *LoadContactsService {
	return &LoadContactsService{
		s3Client:    s3Client,
		bucket:      bucket,
		kafkaWriter: kafkaWriter,
	}
}

// ProcessUpload streams the payload to S3, generates a unique storage key, and publishes a
// LoadContactsTask message to Kafka.
func (lcs *LoadContactsService) ProcessUpload(ctx context.Context, userID int, filename string, payload io.ReadSeeker) error {
	key := fmt.Sprintf("contacts/%d_%s", time.Now().UnixNano(), filename)

	_, err := lcs.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(lcs.bucket),
		Key:    aws.String(key),
		Body:   payload,
	})
	if err != nil {
		return err
	}

	jsonTask, err := json.Marshal(&domain.LoadContactsTask{
		UserID: userID,
		S3Key:  key,
	})
	if err != nil {
		return err
	}

	err = lcs.kafkaWriter.WriteMessages(ctx, kafka.Message{
		Value: jsonTask,
	})
	if err != nil {
		return err
	}

	return nil
}
