package observability

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/domain"
)

// TasksCreatedTotal は Task 作成結果の分類カウンタ。
// outcome ラベル値: "success" / "db_error" / "unknown"
var TasksCreatedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "go_web_textbook",
		Subsystem: "tasks",
		Name:      "created_total",
		Help:      "Number of task create attempts.",
	},
	[]string{"outcome"},
)

// TaskOperationDuration は CRUD 別のヒストグラム。
// operation ラベル値: "create" / "read" / "update" / "delete"
var TaskOperationDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "go_web_textbook",
		Subsystem: "tasks",
		Name:      "operation_duration_seconds",
		Help:      "Duration of task CRUD operations.",
		Buckets: []float64{
			0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5,
		},
	},
	[]string{"operation"},
)

// InitSLI は default registry に SLI メトリクスを登録する。
// /metrics は default registry を読むため、これ呼出しだけで自動露出される。
// 重複登録は panic するため main からちょうど 1 回だけ呼ぶ。
func InitSLI() {
	prometheus.MustRegister(TasksCreatedTotal, TaskOperationDuration)
}

// OutcomeFor は usecase 層のエラーを Prometheus ラベル値に分類する。
// errors.Is で判定し、ドメイン DB エラーは "db_error"、未分類は "unknown"。
func OutcomeFor(err error) string {
	switch {
	case errors.Is(err, domain.ErrTaskNotFound),
		errors.Is(err, domain.ErrDuplicate),
		errors.Is(err, domain.ErrForeignKey),
		errors.Is(err, domain.ErrCheckViolation):
		return "db_error"
	default:
		return "unknown"
	}
}
