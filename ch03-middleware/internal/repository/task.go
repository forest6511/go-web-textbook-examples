package repository

import (
	"context"
	"sync"

	"github.com/forest6511/go-web-textbook-examples/ch03-middleware/internal/domain"
)

type InMemoryTaskRepo struct {
	mu     sync.Mutex
	nextID int64
	store  map[int64]*domain.Task
}

func NewInMemoryTaskRepo() *InMemoryTaskRepo {
	return &InMemoryTaskRepo{store: map[int64]*domain.Task{}}
}

func (r *InMemoryTaskRepo) Create(_ context.Context, t *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	t.ID = r.nextID
	r.store[t.ID] = t
	return nil
}

func (r *InMemoryTaskRepo) Get(_ context.Context, id int64) (*domain.Task, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.store[id]
	return t, ok
}

func (r *InMemoryTaskRepo) List(
	_ context.Context, status string, limit int,
) []*domain.Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*domain.Task, 0, limit)
	for _, t := range r.store {
		if status != "all" && t.Status != status {
			continue
		}
		out = append(out, t)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func (r *InMemoryTaskRepo) Update(
	_ context.Context, id int64, fn func(*domain.Task),
) (*domain.Task, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.store[id]
	if !ok {
		return nil, false
	}
	fn(t)
	return t, true
}

func (r *InMemoryTaskRepo) Delete(_ context.Context, id int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return false
	}
	delete(r.store, id)
	return true
}
