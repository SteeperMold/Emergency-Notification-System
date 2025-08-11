package bootstrap

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// NewS3Client creates and returns an S3 client configured with static credentials
// and custom endpoint options. It fatally logs if session creation fails.
func NewS3Client(s3config *S3Config) *s3.S3 {
	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(s3config.ID, s3config.Key, ""),
		Region:           aws.String(s3config.Region),
		Endpoint:         aws.String(s3config.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		log.Fatalf("failed to connect to S3: %v", err)
	}

	svc := s3.New(sess)

	return svc
}
