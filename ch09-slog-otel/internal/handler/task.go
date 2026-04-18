package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/apperror"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/domain"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/usecase"
)

type TaskHandler struct {
	usecase *usecase.TaskUsecase
}

func NewTaskHandler(u *usecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{usecase: u}
}

type CreateTaskRequest struct {
	Title string `json:"title" binding:"required,max=100"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open in_progress done"`
}

type TaskIDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

// devUserID は Ch 07 で認証を導入するまでの開発用固定ユーザー ID
const devUserID int64 = 1

// Create は新規タスクを作成する
func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	task, err := h.usecase.Create(c.Request.Context(), devUserID, req.Title)
	if err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	c.Header("Location", "/api/v1/tasks/"+strconv.FormatInt(task.ID, 10))
	c.JSON(http.StatusCreated, task)
}

// List はタスク一覧を返す。?limit=20&offset=0 のようなページングを受ける
func (h *TaskHandler) List(c *gin.Context) {
	limit, offset, err := parseListQuery(c)
	if err != nil {
		_ = c.Error(apperror.NewBadRequest(err.Error(), err))
		return
	}

	tasks, err := h.usecase.List(c.Request.Context(), devUserID, limit, offset)
	if err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func parseListQuery(c *gin.Context) (int32, int32, error) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		return 0, 0, errors.New("limit must be 1..100")
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		return 0, 0, errors.New("offset must be >= 0")
	}
	return int32(limit), int32(offset), nil
}

// Get は ID で特定したタスクを返す
func (h *TaskHandler) Get(c *gin.Context) {
	var p TaskIDParam
	if err := c.ShouldBindUri(&p); err != nil {
		_ = c.Error(err)
		return
	}

	task, err := h.usecase.Get(c.Request.Context(), devUserID, p.ID)
	if err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	c.JSON(http.StatusOK, task)
}

// UpdateStatus はステータスのみ更新する。本章では単一フィールド更新に限定
func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	var p TaskIDParam
	if err := c.ShouldBindUri(&p); err != nil {
		_ = c.Error(err)
		return
	}
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	err := h.usecase.UpdateStatus(
		c.Request.Context(), devUserID, p.ID, domain.Status(req.Status),
	)
	if err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	c.Status(http.StatusNoContent)
}

// Delete は ID で特定したタスクを削除する
func (h *TaskHandler) Delete(c *gin.Context) {
	var p TaskIDParam
	if err := c.ShouldBindUri(&p); err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.usecase.Delete(c.Request.Context(), devUserID, p.ID); err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	c.Status(http.StatusNoContent)
}
