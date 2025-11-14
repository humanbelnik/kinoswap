package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

func baseURL() string {
	env := os.Getenv("ENV")
	switch env {
	case "CI":
		return "http://core-app:8080/api/v1"
	}
	return "http://localhost:8080/api/v1"
}

const (
	adminCode = "shared"
)

type AuthRequest struct {
	Code string `json:"code"`
}

type MovieRequest struct {
	Title    string   `json:"title"`
	Year     int      `json:"year"`
	Rating   float64  `json:"rating"`
	Genres   []string `json:"genres"`
	Overview string   `json:"overview"`
}

func main() {
	fmt.Println("Starting E2E tests for Kinoswap API...")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if !waitForService(client) {
		os.Exit(1)
	}

	adminToken, err := authenticate(client)
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Authenticated successfully. Token: %s...\n", adminToken[:10])

	movieID, err := createMovie(client, adminToken)
	if err != nil {
		fmt.Printf("Create movie failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Movie created successfully. ID: %s\n", movieID)

	moviesCount, err := getMovies(client, adminToken)
	if err != nil {
		fmt.Printf("Get movies failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Retrieved %d movies successfully\n", moviesCount)

	fmt.Println("\n All E2E tests passed!")
}

func waitForService(client *http.Client) bool {
	fmt.Println(" Waiting for service to be ready...")

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(baseURL() + "/movies")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
				fmt.Println(" Service is ready!")
				return true
			}
		}

		if i < maxRetries-1 {
			fmt.Printf(" Service not ready yet (attempt %d/%d)...\n", i+1, maxRetries)
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Println(" Service didn't start in time")
	return false
}

func authenticate(client *http.Client) (string, error) {
	fmt.Println("\n Step 1: Authenticating...")

	authReq := AuthRequest{
		Code: adminCode,
	}

	body, err := json.Marshal(authReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth request: %v", err)
	}

	resp, err := client.Post(baseURL()+"/auth", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("auth request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("auth returned status %d: %s", resp.StatusCode, string(body))
	}

	token := resp.Header.Get("X-admin-token")
	if token == "" {
		return "", fmt.Errorf("admin token not found in response headers")
	}

	return token, nil
}

func createMovie(client *http.Client, adminToken string) (string, error) {
	fmt.Println("\n Step 2: Creating movie...")

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	movieData := MovieRequest{
		Title:    "Inception",
		Year:     2010,
		Rating:   8.8,
		Genres:   []string{"sci-fi", "action", "thriller"},
		Overview: "A thief who steals corporate secrets through the use of dream-sharing technology.",
	}

	jsonData, err := json.Marshal(movieData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal movie data: %v", err)
	}

	jsonPart, err := writer.CreateFormField("body")
	if err != nil {
		return "", fmt.Errorf("failed to create form field: %v", err)
	}

	_, err = jsonPart.Write(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to write form data: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", baseURL()+"/movies", &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-admin-token", adminToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("movie creation request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create movie returned status %d: %s", resp.StatusCode, string(body))
	}

	return "created", nil
}

func getMovies(client *http.Client, adminToken string) (int, error) {
	fmt.Println("\n Step 3: Getting movies list...")

	req, err := http.NewRequest("GET", baseURL()+"/movies", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-admin-token", adminToken)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("get movies request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("get movies returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Movies []interface{} `json:"movies"`
		Total  int           `json:"total"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to parse movies response: %v", err)
	}

	return response.Total, nil
}
