package router

import (
	"github.com/gin-gonic/gin"

	"github.com/forest6511/go-web-textbook-examples/ch01-gin-intro/internal/handler"
)

// New はルーティング設定済みのエンジンを返す
func New() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", handler.Ping)
	return r
}
