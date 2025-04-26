package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/sjzar/file-store-mcp/pkg/util"
)

// S3Client is a wrapper for the S3 client
type S3Client struct {
	client     *s3.Client
	bucketName string
	region     string
	endpoint   string
	// Add fields for generating signed URLs
	accessKey  string
	secretKey  string
	expiration time.Duration // URL expiration time
}

// S3Config contains configuration for the S3 client
type S3Config struct {
	BucketName  string
	Region      string
	Endpoint    string
	AccessKeyID string
	SecretKey   string
	Session     string
	// Add URL expiration configuration (in seconds)
	URLExpiration int64
}

// NewS3Client creates a new S3 client
func NewS3Client(cfg S3Config) (*S3Client, error) {
	// Configuration options
	var optFns []func(*config.LoadOptions) error

	// Add region configuration
	optFns = append(optFns, config.WithRegion(cfg.Region))
	optFns = append(optFns, config.WithRequestChecksumCalculation(0))
	optFns = append(optFns, config.WithResponseChecksumValidation(0))

	// Add static credentials provider if credentials are provided
	if cfg.AccessKeyID != "" && cfg.SecretKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretKey, cfg.Session),
		))
	}

	// Load configuration
	awsCfg, err := config.LoadDefaultConfig(context.TODO(), optFns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK configuration: %w", err)
	}

	// Create S3 client options
	s3Options := s3.Options{
		Region:      cfg.Region,
		Credentials: awsCfg.Credentials,
	}

	// Use custom endpoint if provided
	if cfg.Endpoint != "" {
		s3Options.BaseEndpoint = aws.String(cfg.Endpoint)
	}

	// Create S3 client
	client := s3.New(s3Options)

	// Set default expiration if not provided
	expiration := time.Hour * 24 * 7 // 7 days default
	if cfg.URLExpiration > 0 {
		expiration = time.Duration(cfg.URLExpiration) * time.Second
	}

	return &S3Client{
		client:     client,
		bucketName: cfg.BucketName,
		region:     cfg.Region,
		endpoint:   cfg.Endpoint,
		accessKey:  cfg.AccessKeyID,
		secretKey:  cfg.SecretKey,
		expiration: expiration,
	}, nil
}

// UploadFile uploads a local file to S3 and returns the download URL
func (s *S3Client) UploadFile(ctx context.Context, path string, filename string) (string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Format the object key using the provided format
	objectKey := filename
	if len(objectKey) == 0 {
		objectKey = uuid.New().String()
	}

	// Upload the file to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(util.GetContentType(filename)),
		// Remove public ACL as it's not supported by many S3 compatible services
		// ACL:         types.ObjectCannedACLPublicRead,
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Generate a presigned URL for the uploaded object
	presignClient := s3.NewPresignClient(s.client)
	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = s.expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
}

// Upload uploads data from an io.Reader to S3 and returns the download URL
func (s *S3Client) Upload(ctx context.Context, body io.Reader, filename string) (string, error) {
	// Format the object key using the provided format
	objectKey := filename
	if len(objectKey) == 0 {
		objectKey = uuid.New().String()
	}

	// Upload the data to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(objectKey),
		Body:        body,
		ContentType: aws.String(util.GetContentType(filename)),
		// Remove public ACL as it's not supported by many S3 compatible services
		// ACL:         types.ObjectCannedACLPublicRead,
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload data to S3: %w", err)
	}

	// Generate a presigned URL for the uploaded object
	presignClient := s3.NewPresignClient(s.client)
	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = s.expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
}
