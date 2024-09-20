package storage

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"news-feed/pkg/logger"
	"time"
)

type MinioStorageInterface interface {
	UploadFile(fileName string, file io.Reader) (string, error)
	GenerateFileURL(fileName string) (string, error)
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
			Creds:  credentials.NewStatic(accessKeyID, secretAccessKey, "", credentials.SignatureDefault),
			Secure: false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not initialize MinIO client: %w", err)
	}

	// Check if the bucket exists
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if !exists {
		// Create the bucket if it doesn't exist
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		} else {
			fmt.Printf("Successfully created bucket: %s\n", bucketName)
		}
	} else {
		fmt.Printf("Bucket %s already exists.\n", bucketName)
	}

	return &MinioStorage{
		client: minioClient,
		bucket: bucketName,
	}, nil
}

func (s *MinioStorage) UploadFile(fileName string, file io.Reader) (string, error) {
	// Upload the file to MinIO
	_, err := s.client.PutObject(context.Background(), s.bucket, fileName, file, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("could not upload file: %w", err)
	}

	fileURL := fmt.Sprintf("%s/%s/%s", s.client.EndpointURL(), s.bucket, fileName)
	return fileURL, nil
}

func (s *MinioStorage) GenerateFileURL(fileName string) (string, error) {
	expires := time.Minute * 15
	preSignedURL, err := s.client.PresignedPutObject(context.Background(), s.bucket, fileName, expires)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error when generate pre signed url %v", err))
		return "", err
	}
	return preSignedURL.String(), nil
}

func (s *MinioStorage) GetFileURL(fileName string) string {
	// Generate a pre-signed URL for the file
	objectURL := fmt.Sprintf("%s/%s/%s", s.client.EndpointURL(), s.bucket, fileName)
	return objectURL
}
