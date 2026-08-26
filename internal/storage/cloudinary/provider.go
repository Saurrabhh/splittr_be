package cloudinary

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/Saurrabhh/splittr_be/internal/storage"
)

// Provider implements storage.Service using Cloudinary SDK.
type Provider struct {
	cld *cloudinary.Cloudinary
}

// New creates a new Cloudinary Provider instance.
func New(cloudName, apiKey, apiSecret string) (*Provider, error) {
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, errors.New("cloudinary credentials cannot be empty")
	}
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to init cloudinary: %w", err)
	}
	return &Provider{cld: cld}, nil
}

// Upload uploads a file to Cloudinary.
func (p *Provider) Upload(ctx context.Context, params storage.UploadParams) (*storage.UploadResult, error) {
	if params.File == nil {
		return nil, errors.New("file reader cannot be nil")
	}

	uploadParams := uploader.UploadParams{
		Folder: params.Folder,
	}

	if params.Width > 0 && params.Height > 0 {
		uploadParams.Transformation = fmt.Sprintf("w_%d,h_%d,c_fill", params.Width, params.Height)
	}

	resp, err := p.cld.Upload.Upload(ctx, params.File, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return &storage.UploadResult{
		URL:      resp.SecureURL,
		PublicID: resp.PublicID,
		Size:     int64(resp.Bytes),
	}, nil
}

// Delete removes a file from Cloudinary given its public ID.
func (p *Provider) Delete(ctx context.Context, publicID string) error {
	if publicID == "" {
		return errors.New("publicID cannot be empty")
	}
	_, err := p.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}
