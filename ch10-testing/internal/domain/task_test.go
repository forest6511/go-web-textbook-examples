package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forest6511/go-web-textbook-examples/ch10-testing/internal/domain"
)

type newTaskCase struct {
	name    string
	title   string
	wantErr error
}

func TestNewTask(t *testing.T) {
	t.Parallel()
	cases := []newTaskCase{
		{"空文字は拒否", "", domain.ErrTitleRequired},
		{"200 文字は許容", strings.Repeat("あ", 200), nil},
		{"201 文字は拒否", strings.Repeat("あ", 201), domain.ErrTitleTooLong},
		{"正常系", "牛乳を買う", nil},
	}
	runNewTaskCases(t, cases)
}

func runNewTaskCases(t *testing.T, cases []newTaskCase) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.NewTask(1, tc.title)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.title, got.Title)
			assert.Equal(t, domain.StatusOpen, got.Status)
			assert.Equal(t, int64(1), got.UserID)
		})
	}
}
