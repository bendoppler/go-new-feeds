package storage

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type MinioStorageInterface interface {
	UploadFile(ctx context.Context, fileName string, file io.Reader) (string, error)
	GetFileURL(fileName string) string
}

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(endpoint, accessKeyID, secretAccessKey, bucketName string) (*MinioStorage, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(
		endpoint, &minio.Options{
			Creds:  credentials.NewStatic(accessKeyID, secretAccessKey, "", credentials.SignatureAnonymous),
			Secure: false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not initialize MinIO client: %w", err)
	}

	// Check if the bucket exists
	found, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil || !found {
		return nil, fmt.Errorf("bucket %s does not exist: %w", bucketName, err)
	}

	return &MinioStorage{
		client: minioClient,
		bucket: bucketName,
	}, nil
}

func (s *MinioStorage) UploadFile(ctx context.Context, fileName string, file io.Reader) (string, error) {
	// Upload the file to MinIO
	_, err := s.client.PutObject(ctx, s.bucket, fileName, file, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("could not upload file: %w", err)
	}

	fileURL := fmt.Sprintf("https://%s/%s/%s", s.client.EndpointURL(), s.bucket, fileName)
	return fileURL, nil
}

func (s *MinioStorage) GetFileURL(fileName string) string {
	// Generate a pre-signed URL for the file
	objectURL := fmt.Sprintf("https://%s/%s/%s", s.client.EndpointURL(), s.bucket, fileName)
	return objectURL
}
