package handler

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/apperror"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/domain"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/repository"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/storage"
)

const (
	maxUploadSize   = 5 << 20
	mimeSniffLength = 3072
)

var allowedMIME = []string{
	"image/jpeg", "image/png",
	"image/webp", "application/pdf",
}

type AttachmentHandler struct {
	storage storage.ObjectStorage
	repo    repository.AttachmentRepo
}

func NewAttachmentHandler(
	s storage.ObjectStorage, r repository.AttachmentRepo,
) *AttachmentHandler {
	return &AttachmentHandler{storage: s, repo: r}
}

func validateFile(file multipart.File, size int64) (string, error) {
	if size > maxUploadSize {
		return "", apperror.NewPayloadTooLarge(
			"file exceeds 5 MiB", nil)
	}
	mtype, err := mimetype.DetectReader(
		io.LimitReader(file, mimeSniffLength))
	if err != nil {
		return "", apperror.NewBadRequest(
			"cannot detect mime", err)
	}
	if !mimetype.EqualsAny(mtype.String(), allowedMIME...) {
		return "", apperror.NewUnsupportedMediaType(
			"mime not allowed: "+mtype.String(), nil)
	}
	return mtype.String(), nil
}

func openAndValidate(
	header *multipart.FileHeader,
) (multipart.File, string, error) {
	f, err := header.Open()
	if err != nil {
		return nil, "", apperror.NewBadRequest(
			"open upload", err)
	}
	mime, err := validateFile(f, header.Size)
	if err != nil {
		_ = f.Close()
		return nil, "", err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		_ = f.Close()
		return nil, "", apperror.NewInternal(err)
	}
	return f, mime, nil
}

func (h *AttachmentHandler) Upload(c *gin.Context) {
	principal, _ := auth.PrincipalFromContext(c.Request.Context())
	header, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(apperror.NewBadRequest(
			"file field is required", err))
		return
	}
	f, mime, err := openAndValidate(header)
	if err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	defer f.Close()
	att, err := h.persist(c.Request.Context(),
		principal.UserID, header, f, mime)
	if err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	c.JSON(http.StatusCreated, att)
}

func (h *AttachmentHandler) persist(
	ctx context.Context, ownerID int64,
	header *multipart.FileHeader,
	f multipart.File, mime string,
) (domain.Attachment, error) {
	id, _ := uuid.NewV7()
	key := fmt.Sprintf("user/%d/%s%s",
		ownerID, id.String(),
		filepath.Ext(header.Filename))
	if err := h.storage.Put(ctx, key, f,
		storage.PutOptions{ContentType: mime}); err != nil {
		return domain.Attachment{}, err
	}
	return h.repo.Create(ctx, repository.NewAttachment{
		ID: id, OwnerID: ownerID,
		ObjectKey:   key,
		Filename:    filepath.Base(header.Filename),
		ContentType: mime, SizeBytes: header.Size,
	})
}

type presignReq struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	SizeBytes   int64  `json:"size_bytes" binding:"required,gt=0"`
}

func validatePresign(req presignReq) error {
	if req.SizeBytes > 100<<20 {
		return apperror.NewPayloadTooLarge(
			"file exceeds 100 MiB", nil)
	}
	if !mimetype.EqualsAny(req.ContentType, allowedMIME...) {
		return apperror.NewUnsupportedMediaType(
			"mime not allowed", nil)
	}
	return nil
}

func (h *AttachmentHandler) PresignUpload(c *gin.Context) {
	principal, _ := auth.PrincipalFromContext(c.Request.Context())
	var req presignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.NewBadRequest("invalid request", err))
		return
	}
	if err := validatePresign(req); err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	id, _ := uuid.NewV7()
	key := fmt.Sprintf("user/%d/%s%s",
		principal.UserID, id.String(),
		filepath.Ext(req.Filename))
	url, err := h.storage.PresignPut(c.Request.Context(), key,
		15*time.Minute,
		storage.PutOptions{ContentType: req.ContentType})
	if err != nil {
		_ = c.Error(apperror.NewInternal(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"upload_url": url,
		"object_key": key,
		"expires_in": 900,
	})
}

func (h *AttachmentHandler) GetDownloadURL(c *gin.Context) {
	principal, _ := auth.PrincipalFromContext(c.Request.Context())
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apperror.NewBadRequest("invalid id", err))
		return
	}
	att, err := h.repo.GetByID(c.Request.Context(), id, principal.UserID)
	if err != nil {
		_ = c.Error(apperror.FromDomain(err))
		return
	}
	url, err := h.storage.PresignGet(c.Request.Context(),
		att.ObjectKey, 5*time.Minute)
	if err != nil {
		_ = c.Error(apperror.NewInternal(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"download_url": url,
		"expires_in":   300,
	})
}
