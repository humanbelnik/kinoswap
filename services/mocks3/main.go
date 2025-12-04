package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func main() {
	addr := getEnv("MOCK_S3_ADDR", ":9090")
	server := NewMockS3Server(addr)

	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Mock S3 server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Mock S3 server...")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type MockS3Server struct {
	server *http.Server
	data   *sync.Map
}

type ObjectData struct {
	Content  []byte
	Metadata map[string]string
}

func NewMockS3Server(addr string) *MockS3Server {
	mux := http.NewServeMux()
	server := &MockS3Server{
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		data: &sync.Map{},
	}

	mux.HandleFunc("/hbk-test-bucket/*", server.handleRequest)
	return server
}

func (m *MockS3Server) Start() error {
	log.Printf("Mock S3 server starting on %s", m.server.Addr)
	return m.server.ListenAndServe()
}

func (m *MockS3Server) Stop(ctx context.Context) error {
	return m.server.Shutdown(ctx)
}

func (m *MockS3Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	bucket, key := extractBucketAndKey(r.URL.Path)

	switch r.Method {
	case http.MethodHead:
		if key == "" {
			m.headBucket(w, r, bucket)
		} else {
			m.headObject(w, r, bucket, key)
		}
	case http.MethodGet:
		m.getObject(w, r, bucket, key)
	case http.MethodPut:
		m.putObject(w, r, bucket, key)
	case http.MethodDelete:
		m.deleteObject(w, r, bucket, key)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func extractBucketAndKey(path string) (string, string) {
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func (m *MockS3Server) headBucket(w http.ResponseWriter, r *http.Request, bucket string) {
	if bucket == "" {
		http.Error(w, "Bucket name required", http.StatusBadRequest)
		return
	}

	// Always return that bucket exists for simplicity
	w.WriteHeader(http.StatusOK)
}

func (m *MockS3Server) headObject(w http.ResponseWriter, r *http.Request, bucket, key string) {
	objKey := fmt.Sprintf("%s/%s", bucket, key)
	if _, exists := m.data.Load(objKey); !exists {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (m *MockS3Server) getObject(w http.ResponseWriter, r *http.Request, bucket, key string) {
	objKey := fmt.Sprintf("%s/%s", bucket, key)
	obj, exists := m.data.Load(objKey)
	if !exists {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	objectData := obj.(*ObjectData)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(objectData.Content)
}

func (m *MockS3Server) putObject(w http.ResponseWriter, r *http.Request, bucket, key string) {
	fmt.Println("Triggered")
}

func (m *MockS3Server) deleteObject(w http.ResponseWriter, r *http.Request, bucket, key string) {
	objKey := fmt.Sprintf("%s/%s", bucket, key)
	m.data.Delete(objKey)
	w.WriteHeader(http.StatusNoContent)
}
