package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/apperror"
)

type Config struct {
	Region    string
	Bucket    string
	Endpoint  string // MinIO 等。空なら AWS S3
	AccessKey string // ローカル開発のみ
	SecretKey string // ローカル開発のみ
}

func LoadConfig() Config {
	return Config{
		Region:    envOr("AWS_REGION", "ap-northeast-1"),
		Bucket:    envOr("S3_BUCKET", "uploads"),
		Endpoint:  os.Getenv("S3_ENDPOINT"),
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
}

func envOr(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func NewS3Client(ctx context.Context, cfg Config) (*s3.Client, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey, cfg.SecretKey, ""),
		))
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	}), nil
}

type S3Storage struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
}

func NewS3Storage(client *s3.Client, bucket string) *S3Storage {
	return &S3Storage{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    bucket,
	}
}

func (s *S3Storage) Put(ctx context.Context, key string,
	body io.Reader, opts PutOptions) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(opts.ContentType),
		Metadata:    opts.Metadata,
	})
	return err
}

func (s *S3Storage) Get(ctx context.Context, key string,
) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, apperror.NewNotFound("object not found", err)
		}
		return nil, err
	}
	return out.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Storage) PresignPut(ctx context.Context, key string,
	expires time.Duration, opts PutOptions) (string, error) {
	req, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(opts.ContentType),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *S3Storage) PresignGet(ctx context.Context, key string,
	expires time.Duration) (string, error) {
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
