package config

import "fmt"

// S3Config holds AWS S3 connection and upload settings used for storing
// assets such as profile images.
type S3Config struct {
	// Region is the AWS region where the S3 bucket resides.
	// Example: ap-south-1
	Region string

	// AccessKeyID is the AWS access key identifier.
	AccessKeyID string

	// SecretAccessKey is the AWS secret access key.
	SecretAccessKey string

	// Bucket is the name of the S3 bucket.
	Bucket string

	// MaxUploadSizeMB is the maximum allowed file upload size in megabytes.
	MaxUploadSizeMB int
}

// loadS3Config reads AWS S3 settings from environment variables.
func loadS3Config() (S3Config, error) {
	region := getEnv("AWS_REGION", "")
	if region == "" {
		return S3Config{}, fmt.Errorf("AWS_REGION is required but not set")
	}

	accessKey := getEnv("AWS_ACCESS_KEY_ID", "")
	if accessKey == "" {
		return S3Config{}, fmt.Errorf("AWS_ACCESS_KEY_ID is required but not set")
	}

	secretKey := getEnv("AWS_SECRET_ACCESS_KEY", "")
	if secretKey == "" {
		return S3Config{}, fmt.Errorf("AWS_SECRET_ACCESS_KEY is required but not set")
	}

	bucket := getEnv("AWS_S3_BUCKET", "")
	if bucket == "" {
		return S3Config{}, fmt.Errorf("AWS_S3_BUCKET is required but not set")
	}

	maxSizeMB := getEnvInt("PROFILE_IMAGE_MAX_SIZE_MB", 15)
	if maxSizeMB <= 0 {
		return S3Config{}, fmt.Errorf("PROFILE_IMAGE_MAX_SIZE_MB must be a positive integer, got %d", maxSizeMB)
	}

	return S3Config{
		Region:          region,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Bucket:          bucket,
		MaxUploadSizeMB: maxSizeMB,
	}, nil
}
