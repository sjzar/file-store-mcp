package storage

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/sjzar/file-store-mcp/internal/storage/cos"
	"github.com/sjzar/file-store-mcp/internal/storage/empty"
	"github.com/sjzar/file-store-mcp/internal/storage/github"
	"github.com/sjzar/file-store-mcp/internal/storage/oss"
	"github.com/sjzar/file-store-mcp/internal/storage/qiniu"
	"github.com/sjzar/file-store-mcp/internal/storage/s3"
)

// Storage defines the interface for storage services
type Storage interface {
	UploadFile(ctx context.Context, path string) (string, error)
}

// Storage type constants
const (
	StorageTypeEmpty  = "empty"
	StorageTypeS3     = "s3"
	StorageTypeOSS    = "oss"
	StorageTypeCOS    = "cos"
	StorageTypeQiniu  = "qiniu"
	StorageTypeGitHub = "github"
)

// InitStorage initializes a storage service based on environment variables
func InitStorage() Storage {

	// Get storage type, default to empty
	storageType := getEnv("FSM_STORAGE_TYPE", StorageTypeEmpty)

	// Initialize the appropriate storage service based on type
	switch strings.ToLower(storageType) {
	case StorageTypeS3:
		return initS3Storage()
	case StorageTypeOSS:
		return initOSSStorage()
	case StorageTypeCOS:
		return initCOSStorage()
	case StorageTypeQiniu:
		return initQiniuStorage()
	case StorageTypeGitHub:
		return initGitHubStorage()
	case StorageTypeEmpty:
		fallthrough
	default:
		log.Debug().Str("type", storageType).Msg("Using empty storage")
		return empty.New("")
	}
}

// initS3Storage initializes AWS S3 storage service
func initS3Storage() Storage {
	cfg := s3.S3Config{
		BucketName:    getEnv("FSM_S3_BUCKET", ""),
		Region:        getEnv("FSM_S3_REGION", ""),
		Endpoint:      getEnv("FSM_S3_ENDPOINT", ""),
		AccessKeyID:   getEnv("FSM_S3_ACCESS_KEY", ""),
		SecretKey:     getEnv("FSM_S3_SECRET_KEY", ""),
		Session:       getEnv("FSM_S3_SESSION", ""),
		URLExpiration: getEnvInt64("FSM_S3_URL_EXPIRATION", 604800), // Default 7 days (in seconds)
	}

	client, err := s3.NewS3Client(cfg)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to initialize S3 storage, falling back to empty storage")
		return empty.New(err.Error())
	}
	log.Debug().Str("bucket", cfg.BucketName).Str("region", cfg.Region).Msg("S3 storage initialized")
	return client
}

// initOSSStorage initializes Aliyun OSS storage service
func initOSSStorage() Storage {
	cfg := oss.OSSConfig{
		Endpoint:        getEnv("FSM_OSS_ENDPOINT", ""),
		AccessKeyID:     getEnv("FSM_OSS_ACCESS_KEY", ""),
		AccessKeySecret: getEnv("FSM_OSS_SECRET_KEY", ""),
		BucketName:      getEnv("FSM_OSS_BUCKET", ""),
		Domain:          getEnv("FSM_OSS_DOMAIN", ""),
		URLExpiration:   getEnvInt64("FSM_OSS_URL_EXPIRATION", 604800), // Default 7 days (in seconds)
	}

	client, err := oss.NewOSSClient(cfg)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to initialize Aliyun OSS storage, falling back to empty storage")
		return empty.New(err.Error())
	}
	log.Debug().Str("bucket", cfg.BucketName).Str("endpoint", cfg.Endpoint).Msg("Aliyun OSS storage initialized")
	return client
}

// initCOSStorage initializes Tencent COS storage service
func initCOSStorage() Storage {
	cfg := cos.COSConfig{
		BucketName:    getEnv("FSM_COS_BUCKET", ""),
		Region:        getEnv("FSM_COS_REGION", ""),
		AppID:         getEnv("FSM_COS_APP_ID", ""),
		SecretID:      getEnv("FSM_COS_ACCESS_KEY", ""),
		SecretKey:     getEnv("FSM_COS_SECRET_KEY", ""),
		Domain:        getEnv("FSM_COS_DOMAIN", ""),
		UseHTTPS:      getEnvBool("FSM_COS_USE_HTTPS", true),
		UseAccelerate: getEnvBool("FSM_COS_USE_ACCELERATE", false),
		URLExpiration: getEnvInt64("FSM_COS_URL_EXPIRATION", 604800), // Default 7 days (in seconds)
	}

	client, err := cos.NewCOSClient(cfg)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to initialize Tencent COS storage, falling back to empty storage")
		return empty.New(err.Error())
	}
	log.Debug().Str("bucket", cfg.BucketName).Str("region", cfg.Region).Msg("Tencent COS storage initialized")
	return client
}

// initQiniuStorage initializes Qiniu Kodo storage service
func initQiniuStorage() Storage {
	cfg := qiniu.QiniuConfig{
		AccessKey:     getEnv("FSM_QINIU_ACCESS_KEY", ""),
		SecretKey:     getEnv("FSM_QINIU_SECRET_KEY", ""),
		BucketName:    getEnv("FSM_QINIU_BUCKET", ""),
		Domain:        getEnv("FSM_QINIU_DOMAIN", ""),
		Region:        getEnv("FSM_QINIU_REGION", "z0"),                // Default to East China
		URLExpiration: getEnvInt64("FSM_QINIU_URL_EXPIRATION", 604800), // Default 7 days (in seconds)
	}

	client, err := qiniu.NewQiniuClient(cfg)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to initialize Qiniu storage, falling back to empty storage")
		return empty.New(err.Error())
	}
	log.Debug().Str("bucket", cfg.BucketName).Str("region", cfg.Region).Msg("Qiniu storage initialized")
	return client
}

// initGitHubStorage initializes GitHub storage service
func initGitHubStorage() Storage {
	cfg := github.GitHubConfig{
		Token:        getEnv("FSM_GITHUB_TOKEN", ""),
		Owner:        getEnv("FSM_GITHUB_OWNER", ""),
		Repo:         getEnv("FSM_GITHUB_REPO", ""),
		Branch:       getEnv("FSM_GITHUB_BRANCH", "main"),
		Path:         getEnv("FSM_GITHUB_PATH", ""),
		CustomDomain: getEnv("FSM_GITHUB_DOMAIN", ""),
	}

	client, err := github.NewGitHubClient(cfg)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to initialize GitHub storage, falling back to empty storage")
		return empty.New(err.Error())
	}
	log.Debug().Str("owner", cfg.Owner).Str("repo", cfg.Repo).Str("branch", cfg.Branch).Msg("GitHub storage initialized")
	return client
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvBool gets a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.ToLower(value) == "true" || value == "1" || value == "yes"
}

// getEnvInt64 gets an int64 environment variable or returns a default value
func getEnvInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result int64
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
