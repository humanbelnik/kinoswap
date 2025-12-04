package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type RRBalancer struct {
	servers []string
	cur     int
}

func (b *RRBalancer) NextServer() string {
	if len(b.servers) == 0 {
		return ""
	}

	b.cur++
	n := b.cur
	index := (n - 1) % len(b.servers)
	return b.servers[index]
}

type AuthClient interface {
	ValidateToken(token string) (bool, error)
	Authenticate(code string) (string, error)
}

type HTTPAuthClient struct {
	balancer   *RRBalancer
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

func New(serversList string) *HTTPAuthClient {
	servers := make([]string, 0)
	if serversList != "" {
		for _, s := range strings.Split(serversList, ";") {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				servers = append(servers, trimmed)
			}
		}
	}

	return &HTTPAuthClient{
		balancer: &RRBalancer{
			servers: servers},
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: slog.Default(),
	}
}

type ValidateRequest struct {
	Token string `json:"token"`
}

type ValidateResponse struct {
	Valid bool `json:"valid"`
}

func (c *HTTPAuthClient) ValidateToken(token string) (bool, error) {
	reqBody := ValidateRequest{Token: token}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.balancer.servers[c.balancer.cur]+"/validate",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return false, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("auth service returned status: %d", resp.StatusCode)
	}

	var validateResp ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return validateResp.Valid, nil
}

func (c *HTTPAuthClient) Authenticate(code string) (string, error) {
	reqBody := map[string]string{"code": code}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/auth",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("auth failed with status: %d", resp.StatusCode)
	}

	token := resp.Header.Get("X-admin-token")
	if token == "" {
		return "", fmt.Errorf("no token in response")
	}

	return token, nil
}
