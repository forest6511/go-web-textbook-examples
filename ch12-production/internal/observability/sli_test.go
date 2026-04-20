package observability

import (
	"errors"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/domain"
)

func TestOutcomeFor_DomainDBErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"ErrTaskNotFound → db_error", domain.ErrTaskNotFound, "db_error"},
		{"ErrDuplicate → db_error", domain.ErrDuplicate, "db_error"},
		{"ErrForeignKey → db_error", domain.ErrForeignKey, "db_error"},
		{"ErrCheckViolation → db_error", domain.ErrCheckViolation, "db_error"},
		{"wrapped ErrDuplicate → db_error",
			fmt.Errorf("insert task: %w", domain.ErrDuplicate), "db_error"},
		{"unclassified → unknown",
			errors.New("connection refused"), "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := OutcomeFor(tc.err)
			if got != tc.want {
				t.Errorf("OutcomeFor(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

// TestSLIMetrics_Labels は SLI カウンタ / ヒストグラムに期待ラベルで値を入れられるかを検証する。
// promauto ではなく手動登録（InitSLI は main で呼ぶ）なので、テストはローカル登録せず
// 観測面だけを確認する（複数 test binary が同時登録すると panic するのを避けるため）。
func TestSLIMetrics_Labels(t *testing.T) {
	TasksCreatedTotal.Reset()
	TaskOperationDuration.Reset()

	TasksCreatedTotal.WithLabelValues("success").Inc()
	TasksCreatedTotal.WithLabelValues("db_error").Inc()
	TasksCreatedTotal.WithLabelValues("db_error").Inc()

	if v := testutil.ToFloat64(TasksCreatedTotal.WithLabelValues("success")); v != 1 {
		t.Errorf("success counter = %v, want 1", v)
	}
	if v := testutil.ToFloat64(TasksCreatedTotal.WithLabelValues("db_error")); v != 2 {
		t.Errorf("db_error counter = %v, want 2", v)
	}

	TaskOperationDuration.WithLabelValues("create").Observe(0.05)
	TaskOperationDuration.WithLabelValues("delete").Observe(0.12)
	// Histogram の個別値は取り出しにくいので、Collect 側では panic しないことだけ確認
}
