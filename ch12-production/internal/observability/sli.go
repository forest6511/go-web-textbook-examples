package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SLI は Ch 12 の SLO 定義で参照する Prometheus メトリクス群。
// OTel の Meter 経由ではなく prometheus/client_golang を直接使う（Ch 12 本文参照）。
// 理由: Alerting Policy や Grafana ダッシュボードで扱う
//   PromQL / Histogram バケットの制御を明示したいため。
type SLI struct {
	TasksCreatedTotal     *prometheus.CounterVec
	TaskOperationDuration *prometheus.HistogramVec
}

// NewSLI は global な default registry に登録された SLI メトリクスを返す。
// promhttp.Handler() は default registry を参照するため、/metrics に自動露出される。
func NewSLI() *SLI {
	return &SLI{
		TasksCreatedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "app",
				Subsystem: "tasks",
				Name:      "created_total",
				Help:      "Total number of tasks created, partitioned by result.",
			},
			[]string{"result"}, // "success" | "error"
		),
		TaskOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "app",
				Subsystem: "tasks",
				Name:      "operation_duration_seconds",
				Help:      "Duration of task operations in seconds.",
				// 10ms 〜 5 秒の対数バケット。p50/p95/p99 の算出に適切
				Buckets: []float64{
					0.01, 0.025, 0.05, 0.1, 0.25,
					0.5, 1.0, 2.5, 5.0,
				},
			},
			[]string{"op", "result"}, // op: "create"/"list"/"get"/"update"/"delete"
		),
	}
}
