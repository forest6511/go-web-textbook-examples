package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbgen "github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/db/gen"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/domain"
	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/repository"
)

func TestTaskRepo_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now().UTC()
	setupCreateExpectation(mock, now)

	repo := repository.NewTaskRepo(dbgen.New(mock))
	got, err := repo.Create(context.Background(), domain.Task{
		UserID: 1,
		Title:  "牛乳を買う",
		Status: domain.StatusOpen,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(42), got.ID)
	assert.Equal(t, "牛乳を買う", got.Title)
	assert.Equal(t, domain.StatusOpen, got.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// setupCreateExpectation は SQL 期待値の宣言だけを分離。
// テスト関数側はアサートの意図に集中できる。
func setupCreateExpectation(mock pgxmock.PgxPoolIface, now time.Time) {
	mock.ExpectQuery("INSERT INTO tasks").
		WithArgs(int64(1), "牛乳を買う", string(domain.StatusOpen)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "status", "created_at", "updated_at",
		}).AddRow(
			int64(42), int64(1), "牛乳を買う",
			string(domain.StatusOpen), now, now,
		))
}

func TestTaskRepo_Delete_NotFoundMapsToDomainErr(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	// DeleteTask は :execrows → Exec 呼び出し
	// 0 行削除は ErrTaskNotFound に翻訳されるのが Repository の契約
	mock.ExpectExec("DELETE FROM tasks").
		WithArgs(int64(99), int64(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repo := repository.NewTaskRepo(dbgen.New(mock))
	err = repo.Delete(context.Background(), 1, 99)
	require.ErrorIs(t, err, domain.ErrTaskNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepo_Delete_OneRowAffected(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec("DELETE FROM tasks").
		WithArgs(int64(10), int64(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := repository.NewTaskRepo(dbgen.New(mock))
	err = repo.Delete(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
