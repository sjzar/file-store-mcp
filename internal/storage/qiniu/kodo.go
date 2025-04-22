package qiniu

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"

	"github.com/sjzar/file-store-mcp/pkg/util"
)

// QiniuClient is a wrapper for the Qiniu cloud storage client
type QiniuClient struct {
	accessKey  string
	secretKey  string
	bucketName string
	domain     string
	region     string
	expiration time.Duration // URL expiration time
}

// QiniuConfig contains configuration for the Qiniu cloud storage client
type QiniuConfig struct {
	AccessKey     string
	SecretKey     string
	BucketName    string
	Domain        string // Required, Qiniu requires a custom domain for access
	Region        string // Storage region, e.g. "z0"(East China), "z1"(North China), "z2"(South China), "na0"(North America), "as0"(Southeast Asia)
	URLExpiration int64  // URL expiration time in seconds
}

// NewQiniuClient creates a new Qiniu cloud storage client
func NewQiniuClient(cfg QiniuConfig) (*QiniuClient, error) {
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("AccessKey and SecretKey cannot be empty")
	}

	if cfg.BucketName == "" {
		return nil, fmt.Errorf("BucketName cannot be empty")
	}

	if cfg.Domain == "" {
		return nil, fmt.Errorf("domain cannot be empty, Qiniu requires a custom domain for access")
	}

	// Ensure domain format is correct
	domain := cfg.Domain
	if domain[len(domain)-1] == '/' {
		domain = domain[:len(domain)-1]
	}
	if len(domain) > 0 && domain[0:4] != "http" {
		domain = "http://" + domain
	}

	// Set default expiration if not provided
	expiration := time.Hour * 24 * 7 // 7 days default
	if cfg.URLExpiration > 0 {
		expiration = time.Duration(cfg.URLExpiration) * time.Second
	}

	return &QiniuClient{
		accessKey:  cfg.AccessKey,
		secretKey:  cfg.SecretKey,
		bucketName: cfg.BucketName,
		domain:     domain,
		region:     cfg.Region,
		expiration: expiration,
	}, nil
}

// UploadFile uploads a local file to Qiniu cloud and returns the download URL
func (q *QiniuClient) UploadFile(ctx context.Context, path string) (string, error) {
	// Get the filename as the object key
	fileName := filepath.Base(path)

	// Generate a unique object key to avoid filename conflicts
	// Using timestamp as prefix
	objectKey := fmt.Sprintf("%d/%s", time.Now().Unix(), fileName)

	// Create authentication information
	mac := qbox.NewMac(q.accessKey, q.secretKey)

	// Create storage configuration
	cfg := storage.Config{}

	// Set storage region
	switch q.region {
	case "z0":
		cfg.Zone = &storage.ZoneHuadong
	case "z1":
		cfg.Zone = &storage.ZoneHuabei
	case "z2":
		cfg.Zone = &storage.ZoneHuanan
	case "na0":
		cfg.Zone = &storage.ZoneBeimei
	case "as0":
		cfg.Zone = &storage.ZoneXinjiapo
	default:
		// Default to East China region
		cfg.Zone = &storage.ZoneHuadong
	}

	// Use HTTPS
	cfg.UseHTTPS = true
	// Use CDN acceleration
	cfg.UseCdnDomains = true

	// Create form uploader object
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}

	// Create upload policy
	putPolicy := storage.PutPolicy{
		Scope: q.bucketName + ":" + objectKey,
	}
	upToken := putPolicy.UploadToken(mac)

	// Create upload options
	putExtra := storage.PutExtra{
		Params: map[string]string{
			"x:name": fileName,
		},
		MimeType: util.GetContentType(fileName),
	}

	// Upload file
	err := formUploader.PutFile(ctx, &ret, upToken, objectKey, path, &putExtra)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Qiniu cloud: %w", err)
	}

	// Build file download URL with authentication
	downloadURL := storage.MakePrivateURL(mac, q.domain, ret.Key, time.Now().Add(q.expiration).Unix())

	return downloadURL, nil
}
