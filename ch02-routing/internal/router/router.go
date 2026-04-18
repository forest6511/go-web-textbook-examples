package router

import (
	"github.com/gin-gonic/gin"

	"github.com/forest6511/go-web-textbook-examples/ch02-routing/internal/handler"
)

func New(taskHandler *handler.TaskHandler) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskHandler.Create)
			tasks.GET("", taskHandler.List)
			tasks.GET("/:id", taskHandler.Get)
			tasks.PATCH("/:id", taskHandler.Update)
			tasks.DELETE("/:id", taskHandler.Delete)
		}
	}

	return r
}
