package domain

import (
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID          uuid.UUID `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	ObjectKey   string    `json:"object_key"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	UploadedAt  time.Time `json:"uploaded_at"`
}
