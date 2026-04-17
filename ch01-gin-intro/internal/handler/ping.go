package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping は /ping の GET ハンドラ。ヘルスチェックや疎通確認で使う
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
