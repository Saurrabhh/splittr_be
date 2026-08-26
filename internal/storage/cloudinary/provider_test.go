package cloudinary_test

import (
	"context"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/storage"
	"github.com/Saurrabhh/splittr_be/internal/storage/cloudinary"
)

func TestCloudinaryProvider_Validation(t *testing.T) {
	t.Run("Empty Credentials Return Error", func(t *testing.T) {
		_, err := cloudinary.New("", "", "")
		if err == nil {
			t.Error("expected error for empty credentials, got nil")
		}
	})

	t.Run("Empty Upload Params Return Error", func(t *testing.T) {
		provider, err := cloudinary.New("dummy_cloud", "dummy_key", "dummy_secret")
		if err != nil {
			t.Fatalf("expected no error on initialization, got: %v", err)
		}

		_, err = provider.Upload(context.Background(), storage.UploadParams{})
		if err == nil {
			t.Error("expected error when uploading empty file, got nil")
		}
	})
}
