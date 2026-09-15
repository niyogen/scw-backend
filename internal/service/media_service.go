package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"delivery-backend/internal/config"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type MediaService interface {
	UploadPhoto(ctx context.Context, file io.Reader, originalFilename string, contentType string) (string, string, error)
	GetMedia(ctx context.Context, objectPath string) (io.ReadCloser, string, error)
}

type mediaService struct {
	cfg        *config.Config
	gcsClient  *storage.Client
	bucketName string
	localDir   string
}

func NewMediaService(cfg *config.Config) MediaService {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	svc := &mediaService{
		cfg:        cfg,
		bucketName: cfg.GCSBucketName,
		localDir:   "./uploads",
	}

	// Try initializing Google Cloud Storage client with Application Default Credentials
	client, err := storage.NewClient(ctx)
	if err == nil {
		svc.gcsClient = client
		log.Printf("Google Cloud Storage client initialized successfully for bucket: %s", cfg.GCSBucketName)
	} else {
		log.Printf("Google Cloud Storage client not initialized (%v), using local fallback storage directory: %s", err, svc.localDir)
		_ = os.MkdirAll(filepath.Join(svc.localDir, "avatars"), 0755)
	}

	return svc
}

func (s *mediaService) UploadPhoto(ctx context.Context, file io.Reader, originalFilename string, contentType string) (string, string, error) {
	// Clean extension
	ext := filepath.Ext(originalFilename)
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".jpg"
		}
	}

	// Generate safe unique filename
	uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), strings.ToLower(ext))
	objectPath := fmt.Sprintf("avatars/%s", uniqueName)

	if s.gcsClient != nil && s.bucketName != "" {
		// Upload directly to Google Cloud Storage
		bucket := s.gcsClient.Bucket(s.bucketName)
		obj := bucket.Object(objectPath)
		writer := obj.NewWriter(ctx)
		writer.ContentType = contentType
		writer.CacheControl = "public, max-age=31536000"

		if _, err := io.Copy(writer, file); err != nil {
			_ = writer.Close()
			return "", "", fmt.Errorf("failed writing to GCS: %w", err)
		}

		if err := writer.Close(); err != nil {
			return "", "", fmt.Errorf("failed closing GCS writer: %w", err)
		}

		mediaURL := fmt.Sprintf("%s/api/v1/media/%s", strings.TrimRight(s.cfg.AppBaseURL, "/"), objectPath)
		return mediaURL, objectPath, nil
	}

	// Fallback to local storage
	targetPath := filepath.Join(s.localDir, objectPath)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", "", fmt.Errorf("failed creating directory: %w", err)
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return "", "", fmt.Errorf("failed creating local file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", "", fmt.Errorf("failed saving local file: %w", err)
	}

	mediaURL := fmt.Sprintf("%s/api/v1/media/%s", strings.TrimRight(s.cfg.AppBaseURL, "/"), objectPath)
	return mediaURL, objectPath, nil
}

func (s *mediaService) GetMedia(ctx context.Context, objectPath string) (io.ReadCloser, string, error) {
	// Security: clean path
	cleanPath := filepath.Clean(objectPath)
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	if s.gcsClient != nil && s.bucketName != "" {
		bucket := s.gcsClient.Bucket(s.bucketName)
		obj := bucket.Object(cleanPath)

		attrs, err := obj.Attrs(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("object not found on GCS: %w", err)
		}

		reader, err := obj.NewReader(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed creating GCS reader: %w", err)
		}

		return reader, attrs.ContentType, nil
	}

	// Fallback local file
	localFilePath := filepath.Join(s.localDir, cleanPath)
	file, err := os.Open(localFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("local file not found: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(localFilePath))
	var cType string
	switch ext {
	case ".jpg", ".jpeg":
		cType = "image/jpeg"
	case ".png":
		cType = "image/png"
	case ".webp":
		cType = "image/webp"
	case ".gif":
		cType = "image/gif"
	default:
		cType = "application/octet-stream"
	}

	return file, cType, nil
}
