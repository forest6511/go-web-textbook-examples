package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/forest6511/go-web-textbook-examples/ch03-middleware/internal/usecase"
)

type TaskHandler struct {
	usecase *usecase.TaskUsecase
}

func NewTaskHandler(u *usecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{usecase: u}
}

type CreateTaskRequest struct {
	Title    string `json:"title"    binding:"required,max=100"`
	Body     string `json:"body"     binding:"max=2000"`
	Priority int    `json:"priority" binding:"gte=0,lte=3"`
}

type UpdateTaskRequest struct {
	Title    *string `json:"title,omitempty"    binding:"omitempty,max=100"`
	Body     *string `json:"body,omitempty"     binding:"omitempty,max=2000"`
	Priority *int    `json:"priority,omitempty" binding:"omitempty,gte=0,lte=3"`
	Status   *string `json:"status,omitempty"   binding:"omitempty,oneof=open done"`
}

type TaskIDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

// Create は新規タスクを作成し、作成リソースの URL を Location に書く
func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := h.usecase.Create(c.Request.Context(), req.Title, req.Body, req.Priority)

	c.Header("Location", "/api/v1/tasks/"+strconv.FormatInt(task.ID, 10))
	c.JSON(http.StatusCreated, task)
}

// List はタスク一覧を返す。?status=open&limit=20 のようなフィルタを受ける
func (h *TaskHandler) List(c *gin.Context) {
	status := c.DefaultQuery("status", "all")
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "limit must be 1..100",
		})
		return
	}

	tasks := h.usecase.List(c.Request.Context(), status, limit)
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// Get は ID で特定したタスクを返す
func (h *TaskHandler) Get(c *gin.Context) {
	var p TaskIDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})
		return
	}

	task, ok := h.usecase.Get(c.Request.Context(), p.ID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "task not found",
		})
		return
	}
	c.JSON(http.StatusOK, task)
}

// Update は部分更新を反映する
func (h *TaskHandler) Update(c *gin.Context) {
	var p TaskIDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, ok := h.usecase.Update(
		c.Request.Context(), p.ID,
		req.Title, req.Body, req.Priority, req.Status,
	)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// Delete は ID で特定したタスクを削除する
func (h *TaskHandler) Delete(c *gin.Context) {
	var p TaskIDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if !h.usecase.Delete(c.Request.Context(), p.ID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
