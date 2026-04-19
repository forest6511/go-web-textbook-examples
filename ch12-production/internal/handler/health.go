package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler は healthz / readyz の 2 エンドポイントを提供する。
// Ch 12 の設計:
//   - healthz: プロセス生存。依存チェックなし。Cloud Run liveness 相当
//   - readyz:  依存サービス (DB) への疎通確認。Cloud Run readiness 相当
//
// healthz と readyz を分離する理由: DB 瞬断で liveness が 503 になるとインスタンスが
// 再起動されてしまう。readiness なら LB がトラフィックを外すだけで済む。
type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// Liveness は /healthz。依存チェックなしで常に 200 を返す。
// プロセスの deadlock / panic だけを検知する軽量エンドポイント。
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readiness は /readyz。pgxpool.Ping で DB 接続を確認する。
// Ping タイムアウトは 3 秒。5xx を返すと LB から外れる想定。
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	if err := h.pool.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"reason": "database unreachable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
