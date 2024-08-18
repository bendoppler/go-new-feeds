package storage

import (
	"fmt"
	"os"
)

type StorageFactoryInterface interface {
	CreateMinioStorage() (MinioStorageInterface, error)
}

type StorageFactory struct{}

func (f *StorageFactory) CreateMinioStorage() (MinioStorageInterface, error) {
	// Read MinIO configuration from environment variables
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET")

	// Initialize MinIO client
	minioClient, err := NewMinioStorage(endpoint, accessKeyID, secretAccessKey, bucketName)
	if err != nil {
		return nil, fmt.Errorf("could not initialize MinIO client: %w", err)
	}
	return minioClient, nil
}
