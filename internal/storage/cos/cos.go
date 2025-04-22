package cos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/sjzar/file-store-mcp/pkg/util"
)

// COSClient is a wrapper for the Tencent Cloud COS client
type COSClient struct {
	client     *cos.Client
	bucketName string
	region     string
	appID      string
	domain     string // Custom domain, if any
	secretID   string
	secretKey  string
	expiration time.Duration // URL expiration time
}

// COSConfig contains configuration for the COS client
type COSConfig struct {
	BucketName    string
	Region        string
	AppID         string
	SecretID      string
	SecretKey     string
	Domain        string // Optional, custom domain
	UseHTTPS      bool   // Whether to use HTTPS
	UseAccelerate bool   // Whether to use global acceleration domain
	URLExpiration int64  // URL expiration time in seconds
}

// NewCOSClient creates a new COS client
func NewCOSClient(cfg COSConfig) (*COSClient, error) {
	// Build COS service URL
	var bucketURL *url.URL
	var err error

	if cfg.UseAccelerate {
		// Use global acceleration domain
		bucketURL, err = url.Parse(fmt.Sprintf("https://%s-%s.cos.accelerate.myqcloud.com", cfg.BucketName, cfg.AppID))
	} else {
		// Use standard domain
		scheme := "https"
		if !cfg.UseHTTPS {
			scheme = "http"
		}
		bucketURL, err = url.Parse(fmt.Sprintf("%s://%s-%s.cos.%s.myqcloud.com", scheme, cfg.BucketName, cfg.AppID, cfg.Region))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse COS service URL: %w", err)
	}

	// Create base HTTP client
	baseURL := &cos.BaseURL{BucketURL: bucketURL}

	// Create COS client
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})

	// Set default expiration if not provided
	expiration := time.Hour * 24 * 7 // 7 days default
	if cfg.URLExpiration > 0 {
		expiration = time.Duration(cfg.URLExpiration) * time.Second
	}

	return &COSClient{
		client:     client,
		bucketName: cfg.BucketName,
		region:     cfg.Region,
		appID:      cfg.AppID,
		domain:     cfg.Domain,
		secretID:   cfg.SecretID,
		secretKey:  cfg.SecretKey,
		expiration: expiration,
	}, nil
}

// UploadFile uploads a local file to COS and returns the download URL
func (c *COSClient) UploadFile(ctx context.Context, path string) (string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get the filename as the object key
	fileName := filepath.Base(path)

	// Generate a unique object key to avoid filename conflicts
	// Using timestamp as prefix
	objectKey := fmt.Sprintf("%d/%s", time.Now().Unix(), fileName)

	// Set upload options
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: util.GetContentType(fileName),
		},
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			// Set object access permission to public read
			XCosACL: "public-read",
		},
	}

	// Upload file to COS
	_, err = c.client.Object.Put(ctx, objectKey, file, opt)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to COS: %w", err)
	}

	// Build file download URL
	var downloadURL string
	if c.domain != "" {
		// Use custom domain
		downloadURL = fmt.Sprintf("%s/%s", c.domain, objectKey)
	} else {
		// Generate a presigned URL with expiration
		presignedURL, err := c.client.Object.GetPresignedURL(ctx, http.MethodGet, objectKey, c.secretID, c.secretKey, c.expiration, nil)
		if err != nil {
			return "", fmt.Errorf("failed to generate presigned URL: %w", err)
		}
		downloadURL = presignedURL.String()
	}

	return downloadURL, nil
}
