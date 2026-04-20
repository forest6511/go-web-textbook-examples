package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Sample は c.Next() の前後で何が起きるかを示すための雛形。
// Ch 03 冒頭の「ミドルウェアの基本形」で解説する学習用の例示であり、
// router には登録していない。
func Sample() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // 後続ミドルウェアとハンドラを実行する

		elapsed := time.Since(start)
		_ = elapsed // 例示のため使わない
	}
}
