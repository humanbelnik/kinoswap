package infra_s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/humanbelnik/kinoswap/core/internal/model"
)

func MustEstabilishConn() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	fmt.Println("AWS REGION", cfg.Region)
	if err != nil {
		log.Fatal(err)
	}
	return s3.NewFromConfig(cfg)
}

type S3Storage struct {
	client *s3.Client

	prefix     string
	bucketName string
}

func New(bucketName string, client *s3.Client, prefix string) (*S3Storage, error) {
	storage := S3Storage{
		bucketName: bucketName,
		client:     client,
		prefix:     prefix,
	}

	_, err := storage.client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				log.Printf("Bucket %v is available.\n", bucketName)
				err = nil
			default:
				log.Printf("Either you don't have access to bucket %v or another error occurred. "+
					"Here's what happened: %v\n", bucketName, err)
			}
		}
	} else {
		log.Printf("Bucket %v exists and you already own it.", bucketName)
	}

	return &storage, err
}

func (s *S3Storage) buildKey(paths ...string) string {
	var cleaned []string
	for _, p := range paths {
		clean := strings.ReplaceAll(p, "\\", "")
		clean = strings.ReplaceAll(clean, "/", "")
		cleaned = append(cleaned, clean)
	}
	return path.Join(cleaned...)
}

func (s *S3Storage) getFilename(path string) string {
	return filepath.Base(path)
}

func (s *S3Storage) Save(ctx context.Context, obj *model.Poster, readyKey *string) (string, error) {
	log.Println("[S3]: Save triggered")
	var key string
	if readyKey == nil {
		key = s.buildKey(s.prefix, obj.GetParent(), obj.GetFilename())
	} else {
		key = *readyKey
	}
	if _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucketName,
		Key:    &key,
		Body:   bytes.NewReader(obj.GetContent()),
		ACL:    types.ObjectCannedACLPrivate,
	}); err != nil {
		return "", fmt.Errorf("failed to save object to S3: %w", err)
	}
	return key, nil
}

func (s *S3Storage) Load(ctx context.Context, readyKey string) (*model.Poster, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &readyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load object from S3: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object content: %w", err)
	}

	obj := (&model.Poster{}).NewFromData(data, s.getFilename(readyKey))
	return obj.(*model.Poster), nil
}

func (s *S3Storage) Update(ctx context.Context, key string, obj *model.Poster) error {
	_, err := s.Save(ctx, obj, &key)
	return err
}

func (s *S3Storage) Delete(ctx context.Context, readyKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucketName,
		Key:    &readyKey,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Storage) GeneratePresignedURL(ctx context.Context, rawURL string, ttl time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(rawURL),
	}, s3.WithPresignExpires(ttl))

	if err != nil {
		return "", err
	}

	return req.URL, nil
}
