package storage

import (
	"fmt"
	"news-feed/pkg/config/webApp"
)

type StorageFactoryInterface interface {
	CreateMinioStorage() (MinioStorageInterface, error)
}

type StorageFactory struct{}

func (f *StorageFactory) CreateMinioStorage() (MinioStorageInterface, error) {
	// Read MinIO configuration from environment variables
	cfg := webApp.LoadConfig()
	endpoint := cfg.MinIOEndpoint
	accessKeyID := cfg.MinIOAccessKey
	secretAccessKey := cfg.MinIOSecretKey
	bucketName := cfg.MinIOBucket

	// Initialize MinIO client
	minioClient, err := NewMinioStorage(endpoint, accessKeyID, secretAccessKey, bucketName)
	if err != nil {
		return nil, fmt.Errorf("could not initialize MinIO client: %w", err)
	}
	return minioClient, nil
}
