//go:build integration

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"golang.org/x/time/rate"

	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/db"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/handler"
	mw "github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/middleware"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/repository"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/router"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/storage"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/usecase"
	appval "github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/validator"
)

// sharedDSN は TestMain で確立した testcontainers Postgres の DSN。
// Snapshot/Restore による 1 コンテナ共有パターン（Ch 10 本文参照）。
var sharedDSN string

func TestMain(m *testing.M) {
	if _, err := appval.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "setup validator: %v\n", err)
		os.Exit(1)
	}
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	ctr, err := tcpg.Run(ctx, "postgres:17-alpine",
		tcpg.WithDatabase("app"),
		tcpg.WithUsername("app"),
		tcpg.WithPassword("app"),
		tcpg.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = testcontainers.TerminateContainer(ctr) }()

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection string: %v\n", err)
		os.Exit(1)
	}
	if err := db.RunMigrations(dsn); err != nil {
		fmt.Fprintf(os.Stderr, "run migrations: %v\n", err)
		os.Exit(1)
	}
	// Snapshot は migrate 済みの DB 状態を保存する。各テストは Restore で戻る
	if err := ctr.Snapshot(ctx, tcpg.WithSnapshotName("migrated")); err != nil {
		fmt.Fprintf(os.Stderr, "snapshot: %v\n", err)
		os.Exit(1)
	}
	sharedDSN = dsn

	os.Exit(m.Run())
}

// testServer はテスト用の httptest.Server と pool を返す。
// 各テストの先頭で呼び、t.Cleanup で DB を Snapshot 状態に戻す。
type testServer struct {
	url  string
	pool *pgxpool.Pool
	ts   *httptest.Server
}

func setupServer(t *testing.T) *testServer {
	t.Helper()
	ctx := t.Context()

	// Restore で migrate 済み状態に戻す。データはリークしない
	ctrName := os.Getenv("TEST_CONTAINER_RESTORE_SKIP")
	if ctrName == "" {
		require.NoError(t, restoreSnapshot(ctx))
	}

	pool, err := db.NewPool(ctx, sharedDSN)
	require.NoError(t, err)

	authCfg := &auth.Config{
		HMACSecret:   []byte("0123456789abcdef0123456789abcdef"),
		Issuer:       "test",
		Audience:     "test",
		AccessTTL:    15 * time.Minute,
		RefreshTTL:   7 * 24 * time.Hour,
		Leeway:       30 * time.Second,
		SecureCookie: false,
	}

	taskRepo := repository.NewPostgresTaskRepo(pool)
	txRunner := repository.NewTxRunner(pool)
	taskUc := usecase.New(taskRepo, txRunner)
	th := handler.NewTaskHandler(taskUc)

	userRepo := repository.NewUserRepo(pool)
	refreshRepo := repository.NewRefreshTokenRepo(pool)
	authUc := usecase.NewAuthUsecase(userRepo, refreshRepo, authCfg)
	ah := handler.NewAuthHandler(authUc, authCfg)

	attRepo := repository.NewPgAttachmentRepo(pool)
	memStorage := storage.NewMemoryStorage()
	attH := handler.NewAttachmentHandler(memStorage, attRepo)

	limiter := mw.NewIPRateLimiter(rate.Limit(1000), 2000) // テスト中は広めに
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	r := router.New(router.Deps{
		Logger:            logger,
		RateLimiter:       limiter,
		TaskHandler:       th,
		AuthHandler:       ah,
		AttachmentHandler: attH,
		AuthCfg:           authCfg,
		Production:        false,
	})

	srv := httptest.NewServer(r)
	t.Cleanup(func() {
		srv.Close()
		pool.Close()
	})
	return &testServer{url: srv.URL, pool: pool, ts: srv}
}

// restoreSnapshot は Snapshot 状態に DB を戻す。
// testcontainers-go の Container.Restore はモジュール側で実装されている
var snapshotMu sync.Mutex

func restoreSnapshot(ctx context.Context) error {
	// Snapshot/Restore は testcontainers の pg モジュールが提供。
	// ここでは TestMain で確立した DSN をそのまま使い、各テスト先頭で
	// TRUNCATE によるクリーンアップを行う（Snapshot API の呼び直しは軽量だが、
	// 毎回コンテナ再生成を避けるため本テストでは TRUNCATE で代替する）。
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	pool, err := pgxpool.New(ctx, sharedDSN)
	if err != nil {
		return err
	}
	defer pool.Close()
	_, err = pool.Exec(ctx, `
        TRUNCATE TABLE refresh_tokens, attachments, audits, tasks, users
        RESTART IDENTITY CASCADE
    `)
	return err
}

// loginResponse は signup / login / refresh の JSON 応答
type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// signupAndLogin はテスト用にユーザーを作成し access token と refresh cookie を返す
func (s *testServer) signupAndLogin(t *testing.T, email, password string) (loginResponse, *http.Cookie) {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	req, _ := http.NewRequest(http.MethodPost, s.url+"/api/v1/auth/signup",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var out loginResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))

	var refresh *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == "refresh_token" {
			refresh = ck
			break
		}
	}
	require.NotNil(t, refresh, "refresh_token cookie missing")
	return out, refresh
}

// Ch 04 積み残しバグ: DeleteTask :execrows 化で存在しない ID の DELETE は 404 を返す
func TestDeleteTask_NotFound(t *testing.T) {
	s := setupServer(t)
	login, _ := s.signupAndLogin(t, "user1@example.com", "password123")

	req, _ := http.NewRequest(http.MethodDelete,
		s.url+"/api/v1/tasks/99999", nil)
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes := make([]byte, 512)
	n, _ := resp.Body.Read(bodyBytes)
	t.Logf("response body: %q ct=%s",
		string(bodyBytes[:n]), resp.Header.Get("Content-Type"))

	assert.Equal(t, http.StatusNotFound, resp.StatusCode,
		"存在しない task ID の DELETE は 404 を返すべき")
}

// Ch 04 積み残しバグ確認: 存在する ID は 204 を返す（回帰テスト）
func TestDeleteTask_OK(t *testing.T) {
	s := setupServer(t)
	login, _ := s.signupAndLogin(t, "user2@example.com", "password123")

	// まず作成
	body := `{"title":"delete-me"}`
	req, _ := http.NewRequest(http.MethodPost, s.url+"/api/v1/tasks",
		bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var task struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&task))
	require.NotZero(t, task.ID)

	// 削除
	delReq, _ := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/api/v1/tasks/%d", s.url, task.ID), nil)
	delReq.Header.Set("Authorization", "Bearer "+login.AccessToken)
	delResp, err := http.DefaultClient.Do(delReq)
	require.NoError(t, err)
	defer delResp.Body.Close()
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)

	// 2 回目の削除は 404
	delReq2, _ := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/api/v1/tasks/%d", s.url, task.ID), nil)
	delReq2.Header.Set("Authorization", "Bearer "+login.AccessToken)
	delResp2, err := http.DefaultClient.Do(delReq2)
	require.NoError(t, err)
	defer delResp2.Body.Close()
	assert.Equal(t, http.StatusNotFound, delResp2.StatusCode)
}

// Ch 05 積み残しバグ: 空 body / 途中で切れた body は 500 ではなく 400 を返す
func TestCreateTask_EmptyBody_Returns400(t *testing.T) {
	s := setupServer(t)
	login, _ := s.signupAndLogin(t, "user3@example.com", "password123")

	req, _ := http.NewRequest(http.MethodPost, s.url+"/api/v1/tasks",
		http.NoBody)
	req.Header.Set("Authorization", "Bearer "+login.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"空 body POST は 400 Bad Request を返すべき（io.EOF 振り分け）")
	assert.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))
}

// Ch 07 積み残しバグ: 同じ refresh token の 2 回目 /auth/refresh は 401 token reuse detected。
// ConsumeRefreshToken の UPDATE ... WHERE used_at IS NULL RETURNING の原子性を保証する。
//
// 並行ケース（goroutine 5 本から同時 refresh）は将来の拡張として残す。本テストでは
// 逐次 2 回呼ぶことで「atomic な used_at セット後は必ず 0 rows / pgx.ErrNoRows」経路を検証。
func TestRefresh_ReuseDetection(t *testing.T) {
	s := setupServer(t)
	_, refresh := s.signupAndLogin(t, "user4@example.com", "password123")

	// 1 回目: 成功し新しい refresh token Cookie が返る
	req1, _ := http.NewRequest(http.MethodPost,
		s.url+"/api/v1/auth/refresh", nil)
	req1.AddCookie(refresh)
	resp1, err := http.DefaultClient.Do(req1)
	require.NoError(t, err)
	defer resp1.Body.Close()
	require.Equal(t, http.StatusOK, resp1.StatusCode, "1st refresh should succeed")

	// 2 回目: 同じ元 token を再利用 → ConsumeRefreshToken が 0 rows を返し
	// pgx.ErrNoRows → ErrRefreshReused → 401 token reuse detected
	req2, _ := http.NewRequest(http.MethodPost,
		s.url+"/api/v1/auth/refresh", nil)
	req2.AddCookie(refresh) // 使用済みの古い cookie を再送
	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode,
		"同じ refresh token の 2 回目の使用は 401 reuse detected を返すべき")
	assert.Equal(t, "application/problem+json",
		resp2.Header.Get("Content-Type"))
}
