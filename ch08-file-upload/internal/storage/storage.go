package storage

import (
	"context"
	"io"
	"time"
)

type PutOptions struct {
	ContentType string
	Metadata    map[string]string
}

type ObjectStorage interface {
	Put(ctx context.Context, key string,
		body io.Reader, opts PutOptions) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	PresignPut(ctx context.Context, key string,
		expires time.Duration, opts PutOptions) (string, error)
	PresignGet(ctx context.Context, key string,
		expires time.Duration) (string, error)
}
