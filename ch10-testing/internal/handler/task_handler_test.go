package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/domain"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/router"
	appval "github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/validator"
)

// fakeTaskRepo は TaskRepository の最小実装。
// testify/mock の文字列 API は使わず、Go の型検査に任せる。
type fakeTaskRepo struct {
	created []domain.Task
}

func (r *fakeTaskRepo) Create(
	_ context.Context, t domain.Task,
) (domain.Task, error) {
	t.ID = int64(len(r.created) + 1)
	r.created = append(r.created, t)
	return t, nil
}

func (r *fakeTaskRepo) GetByID(
	_ context.Context, _ int64, id int64,
) (domain.Task, error) {
	return domain.Task{}, domain.ErrTaskNotFound
}

func (r *fakeTaskRepo) ListByUser(
	_ context.Context, _ int64, _, _ int32,
) ([]domain.Task, error) {
	return nil, nil
}

func (r *fakeTaskRepo) UpdateStatus(
	_ context.Context, _, _ int64, _ domain.Status,
) error {
	return nil
}

func (r *fakeTaskRepo) Delete(
	_ context.Context, _, _ int64,
) error {
	return nil
}

func newTestEngine(t *testing.T, repo *fakeTaskRepo) *gin.Engine {
	t.Helper()
	if _, err := appval.Setup(); err != nil {
		t.Fatalf("setup validator: %v", err)
	}
	gin.SetMode(gin.TestMode)
	return router.New(router.Deps{
		TaskRepo: repo,
		AuthSkip: true,
	})
}

func TestCreateTask_OK(t *testing.T) {
	repo := &fakeTaskRepo{}
	engine := newTestEngine(t, repo)

	body, _ := json.Marshal(map[string]string{"title": "牛乳を買う"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.Equal(t, "牛乳を買う", got["title"])
	assert.Len(t, repo.created, 1)
	assert.Equal(t, int64(1), repo.created[0].UserID)
}

func TestCreateTask_MissingTitle_Returns422(t *testing.T) {
	repo := &fakeTaskRepo{}
	engine := newTestEngine(t, repo)

	body, _ := json.Marshal(map[string]string{"title": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	// validator の required タグによるバリデーション失敗 → 422
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Empty(t, repo.created)
}
