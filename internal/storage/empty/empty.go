package empty

import (
	"context"
	"errors"
)

// EmptyStorage is a no-op storage implementation
type EmptyStorage struct {
	Info string // FXIME
}

// New creates a new empty storage instance
func New(info string) *EmptyStorage {
	return &EmptyStorage{
		Info: info,
	}
}

// UploadFile implements the Storage interface but always returns an error
func (e *EmptyStorage) UploadFile(ctx context.Context, path string) (string, error) {
	return "", errors.New("storage service not configured or initialization failed. " + e.Info)
}
