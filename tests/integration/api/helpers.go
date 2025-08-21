package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// HTTPClient provides methods for making HTTP requests in tests
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient creates a new HTTP client for testing
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GET makes a GET request to the specified path
func (c *HTTPClient) GET(path string, headers ...map[string]string) (*http.Response, error) {
	return c.makeRequest("GET", path, nil, headers...)
}

// POST makes a POST request to the specified path with JSON body
func (c *HTTPClient) POST(path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	return c.makeRequest("POST", path, body, headers...)
}

// PUT makes a PUT request to the specified path with JSON body
func (c *HTTPClient) PUT(path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	return c.makeRequest("PUT", path, body, headers...)
}

// DELETE makes a DELETE request to the specified path
func (c *HTTPClient) DELETE(path string, headers ...map[string]string) (*http.Response, error) {
	return c.makeRequest("DELETE", path, nil, headers...)
}

// makeRequest is a helper method to make HTTP requests
func (c *HTTPClient) makeRequest(method, path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set additional headers
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	return c.client.Do(req)
}

// POSTRaw makes a POST request with raw body (for testing invalid JSON)
func (c *HTTPClient) POSTRaw(path string, body string, headers ...map[string]string) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set additional headers
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	return c.client.Do(req)
}

// DecodeResponse decodes a JSON response into the provided interface
func DecodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// ReadResponseBody reads the response body as a string
func ReadResponseBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// GenerateJWTToken generates a JWT token for testing authentication
func GenerateJWTToken(secret string, userID uint, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// AuthHeaders creates authorization headers with JWT token
func AuthHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}
