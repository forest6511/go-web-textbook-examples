package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/forest6511/go-web-textbook-examples/ch08-file-upload/internal/apperror"
)

// MemoryStorage はテスト用のインメモリ実装。
type MemoryStorage struct {
	mu    sync.Mutex
	items map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{items: map[string][]byte{}}
}

func (m *MemoryStorage) Put(_ context.Context, key string,
	body io.Reader, _ PutOptions) error {
	buf, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = buf
	return nil
}

func (m *MemoryStorage) Get(_ context.Context, key string,
) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	buf, ok := m.items[key]
	if !ok {
		return nil, apperror.NewNotFound("object not found",
			errors.New(key))
	}
	return io.NopCloser(bytes.NewReader(buf)), nil
}

func (m *MemoryStorage) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
	return nil
}

func (m *MemoryStorage) PresignPut(_ context.Context, key string,
	_ time.Duration, _ PutOptions) (string, error) {
	return "memory://" + key, nil
}

func (m *MemoryStorage) PresignGet(_ context.Context, key string,
	_ time.Duration) (string, error) {
	return "memory://" + key, nil
}
