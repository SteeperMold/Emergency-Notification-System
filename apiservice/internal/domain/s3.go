package domain

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Client abstracts AWS S3 PutObject operations.
// Implementations of this interface should handle uploading data streams
// to S3 with context-based cancellation and retries.
type S3Client interface {
	PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error)
}
