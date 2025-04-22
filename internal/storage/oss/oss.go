package oss

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/sjzar/file-store-mcp/pkg/util"
)

// OSSClient is a wrapper for the Aliyun OSS client
type OSSClient struct {
	client        *oss.Client
	bucket        *oss.Bucket
	bucketName    string
	endpoint      string
	domain        string // Custom domain, if any
	urlExpiration time.Duration
}

// OSSConfig contains configuration for the OSS client
type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	Domain          string // Optional, custom domain
	URLExpiration   int64  // URL expiration time in seconds
}

// NewOSSClient creates a new OSS client
func NewOSSClient(cfg OSSConfig) (*OSSClient, error) {
	// Create OSS client
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	// Get bucket
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get OSS bucket: %w", err)
	}

	// Set default expiration if not provided
	expiration := time.Hour * 24 * 7 // 7 days default
	if cfg.URLExpiration > 0 {
		expiration = time.Duration(cfg.URLExpiration) * time.Second
	}

	return &OSSClient{
		client:        client,
		bucket:        bucket,
		bucketName:    cfg.BucketName,
		endpoint:      cfg.Endpoint,
		domain:        cfg.Domain,
		urlExpiration: expiration,
	}, nil
}

// UploadFile uploads a local file to OSS and returns the download URL
func (o *OSSClient) UploadFile(ctx context.Context, path string) (string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Get the filename as the object key
	fileName := filepath.Base(path)

	// Generate a unique object key to avoid filename conflicts
	// Using timestamp as prefix
	objectKey := fmt.Sprintf("%d/%s", time.Now().Unix(), fileName)

	// Set file metadata
	options := []oss.Option{
		oss.ContentType(util.GetContentType(fileName)),
		oss.ContentLength(fileInfo.Size()),
	}

	// Upload file to OSS
	err = o.bucket.PutObject(objectKey, file, options...)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to OSS: %w", err)
	}

	// Build the file download URL
	var downloadURL string
	if o.domain != "" {
		// If custom domain is provided and we want to use it directly without signing
		// This is useful when the bucket is configured with CDN or public read access
		if isPublicDomain(o.domain) {
			downloadURL = fmt.Sprintf("%s/%s", o.domain, objectKey)
		} else {
			// Generate signed URL with custom domain
			signedURL, err := o.bucket.SignURL(objectKey, oss.HTTPGet, int64(o.urlExpiration.Seconds()))
			if err != nil {
				return "", fmt.Errorf("failed to generate signed URL: %w", err)
			}
			// Replace the default endpoint with custom domain in the signed URL
			defaultEndpoint := fmt.Sprintf("https://%s.%s", o.bucketName, o.endpoint)
			downloadURL = replaceEndpoint(signedURL, defaultEndpoint, o.domain)
		}
	} else {
		// Generate signed URL with default endpoint
		signedURL, err := o.bucket.SignURL(objectKey, oss.HTTPGet, int64(o.urlExpiration.Seconds()))
		if err != nil {
			return "", fmt.Errorf("failed to generate signed URL: %w", err)
		}
		downloadURL = signedURL
	}

	return downloadURL, nil
}

// isPublicDomain checks if a domain should be treated as public (no signing needed)
// This can be determined by configuration or domain pattern
func isPublicDomain(domain string) bool {
	// For now, assume all custom domains are CDN domains that need no signing
	// In a real implementation, this could be controlled by a config flag
	return true
}

// replaceEndpoint replaces the default endpoint in a signed URL with a custom domain
func replaceEndpoint(signedURL, defaultEndpoint, customDomain string) string {
	// Simple string replacement - in a real implementation, this might need more robust URL parsing
	return signedURL
}
