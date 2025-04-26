package storage

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	Storage Storage
	Config  *Config
}

// NewService creates a new service using environment variables for configuration
func NewService() *Service {
	config := NewConfigFromEnv()
	return &Service{
		Storage: NewStorage(config),
		Config:  config,
	}
}

// NewServiceWithConfig creates a new service using the provided configuration
func NewServiceWithConfig(config *Config) *Service {
	return &Service{
		Storage: NewStorage(config),
		Config:  config,
	}
}

// UploadFile uploads a file to the configured storage service
// Uses the default format or a format specified by environment variable
func (s *Service) UploadFile(ctx context.Context, path string) (string, error) {
	// Get format from environment variable, default to empty string
	format := getEnv("FSM_FILE_FORMAT", "")
	if len(format) == 0 {
		format = "{timestamp}-{filename}{ext}"
	}

	// Get the filename
	filename := filepath.Base(path)

	// Format the object key using the FormatObjectKey function
	formattedFilename := FormatObjectKey(filename, format)

	// Upload the file with the formatted key
	return s.Storage.UploadFile(ctx, path, formattedFilename)
}

// UploadFileWithFormat uploads a file with a custom format string
func (s *Service) UploadFileWithFormat(ctx context.Context, path string, format string) (string, error) {
	if len(format) == 0 {
		format = "{timestamp}-{filename}{ext}"
	}

	// Get the filename
	filename := filepath.Base(path)

	// Format the object key using the FormatObjectKey function
	formattedFilename := FormatObjectKey(filename, format)

	// Upload the file with the formatted key
	return s.Storage.UploadFile(ctx, path, formattedFilename)
}

// Upload uploads data from an io.Reader to the configured storage service
func (s *Service) Upload(ctx context.Context, body io.Reader, filename string) (string, error) {
	// Get format from environment variable, default to empty string
	format := getEnv("FSM_FILE_FORMAT", "")
	if len(format) == 0 {
		format = "{timestamp}-{filename}{ext}"
	}

	// Format the object key using the FormatObjectKey function
	formattedFilename := FormatObjectKey(filename, format)

	// Upload the data with the formatted key
	return s.Storage.Upload(ctx, body, formattedFilename)
}

// UploadWithFormat uploads data from an io.Reader with a custom format string
func (s *Service) UploadWithFormat(ctx context.Context, body io.Reader, filename string, format string) (string, error) {
	if len(format) == 0 {
		format = "{timestamp}-{filename}{ext}"
	}

	// Format the object key using the FormatObjectKey function
	formattedFilename := FormatObjectKey(filename, format)

	// Upload the data with the formatted key
	return s.Storage.Upload(ctx, body, formattedFilename)
}

// FormatObjectKey formats the object key based on the provided format string
// Supports the following placeholders:
// {filename} - original filename without extension
// {ext} - file extension with dot
// {timestamp} - Unix timestamp
// {uuid} - random UUID
// {rand} - random 6-character string
func FormatObjectKey(filename string, format string) string {
	if format == "" {
		// Default format: timestamp/original filename
		return fmt.Sprintf("%d/%s", time.Now().Unix(), filename)
	}

	fileExt := filepath.Ext(filename)
	fileNameWithoutExt := strings.TrimSuffix(filename, fileExt)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	uuidStr := uuid.New().String()

	// Generate random string
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randStr := make([]byte, 6)
	for i := range randStr {
		randStr[i] = charset[rand.Intn(len(charset))]
	}

	// Replace placeholders
	result := format
	result = strings.ReplaceAll(result, "{filename}", fileNameWithoutExt)
	result = strings.ReplaceAll(result, "{ext}", fileExt)
	result = strings.ReplaceAll(result, "{timestamp}", timestamp)
	result = strings.ReplaceAll(result, "{uuid}", uuidStr)
	result = strings.ReplaceAll(result, "{rand}", string(randStr))

	return result
}
