package s3store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type Config struct {
	Bucket    string
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
}

type Store struct {
	client *s3.Client
	bucket string
}

func New(ctx context.Context, input Config) (*Store, error) {
	if input.Bucket == "" {
		return nil, fmt.Errorf("CATALOG_S3_BUCKET is required when storage is enabled")
	}
	if (input.AccessKey == "") != (input.SecretKey == "") {
		return nil, fmt.Errorf("both Catalog S3 access key and secret key are required")
	}
	options := []func(*awsConfig.LoadOptions) error{awsConfig.WithRegion(input.Region)}
	if input.AccessKey != "" {
		options = append(options, awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(input.AccessKey, input.SecretKey, "")))
	}
	cfg, err := awsConfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("configure object storage: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		if input.Endpoint != "" {
			options.BaseEndpoint = aws.String(input.Endpoint)
			options.UsePathStyle = true
		}
	})
	return &Store{client: client, bucket: input.Bucket}, nil
}

func (s *Store) Put(ctx context.Context, key, contentType string, content []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), ContentType: aws.String(contentType), Body: bytes.NewReader(content)})
	return err
}

func (s *Store) Get(ctx context.Context, key string, limit int64) ([]byte, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	content, err := io.ReadAll(io.LimitReader(result.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > limit {
		return nil, fmt.Errorf("stored object exceeds image size limit")
	}
	return content, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	return err
}

// EnsureBucket is used only by the local development bootstrap job.
func (s *Store) EnsureBucket(ctx context.Context) error {
	if _, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)}); err == nil {
		return nil
	}
	_, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)})
	var apiError smithy.APIError
	if errors.As(err, &apiError) && apiError.ErrorCode() == "BucketAlreadyOwnedByYou" {
		return nil
	}
	return err
}
