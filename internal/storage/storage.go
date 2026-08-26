package storage

import (
	"context"
	"io"
)

// UploadParams defines generic parameters required for uploading a file.
type UploadParams struct {
	File        io.Reader
	FileName    string
	ContentType string
	Folder      string
	Width       int
	Height      int
}

// UploadResult defines generic response returned by storage provider after successful upload.
type UploadResult struct {
	URL      string
	PublicID string
	Size     int64
}

// Service defines abstract contract for file storage providers.
type Service interface {
	Upload(ctx context.Context, params UploadParams) (*UploadResult, error)
	Delete(ctx context.Context, publicID string) error
}
