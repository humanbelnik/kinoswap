package infra_s3

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ClientType string

const (
	ClientTypeRealS3 ClientType = "real"
	ClientTypeMock   ClientType = "mock"
)

func MustEstablishConn() *s3.Client {
	clientType := getClientType()

	switch clientType {
	case ClientTypeMock:
		return createMockClient()
	case ClientTypeRealS3:
		fallthrough
	default:
		return createRealClient()
	}
}

func getClientType() ClientType {
	clientType := os.Getenv("S3_CLIENT_TYPE")
	fmt.Println(">>>>>>>>>>>>>", clientType)
	if clientType == string(ClientTypeMock) {
		return ClientTypeMock
	}
	return ClientTypeRealS3
}

func createRealClient() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Using REAL S3 client in region:", cfg.Region)
	return s3.NewFromConfig(cfg)
}

func createMockClient() *s3.Client {
	mockEndpoint := getEnv("MOCK_S3_ENDPOINT", "http://mock-s3-server:9090")

	fmt.Println("Using MOCK S3 client with endpoint:", mockEndpoint)

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               mockEndpoint,
			SigningRegion:     "mock-region",
			HostnameImmutable: true,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("mock", "mock", "")),
		config.WithRegion("mock-region"),
	)
	if err != nil {
		log.Fatal("Failed to create mock S3 config:", err)
	}

	return s3.NewFromConfig(cfg)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
