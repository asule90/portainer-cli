package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/portainer/portainer-cli/pkg/endpoints"
	"github.com/portainer/portainer-cli/pkg/stacks"
)

// APIError represents an error response returned by Portainer API.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.Details != "" && e.Details != e.Message {
		return fmt.Sprintf("portainer API error (status %d): %s - %s", e.StatusCode, e.Message, e.Details)
	}
	if e.Message != "" {
		return fmt.Sprintf("portainer API error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("portainer API error: unexpected status %d", e.StatusCode)
}

// Client is the Portainer API client.
type Client struct {
	baseURL     *url.URL
	httpClient  *http.Client
	apiKey      string
	jwtToken    string
	insecureTLS bool

	stacksService    stacks.Service
	endpointsService endpoints.Service
}

// NewClient creates a new Portainer API Client.
func NewClient(opts ...Option) (*Client, error) {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false, // #nosec G402
		},
	}

	c := &Client{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   60 * time.Second,
		},
	}

	defaultBase, _ := url.Parse("https://localhost:9443")
	c.baseURL = defaultBase

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.stacksService = stacks.NewService(c)
	c.endpointsService = endpoints.NewService(c)
	return c, nil
}

// setBaseURL parses and stores the base URL.
func (c *Client) setBaseURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	// Strip trailing slashes
	u.Path = strings.TrimRight(u.Path, "/")
	c.baseURL = u
	return nil
}

// Stacks returns the stacks service.
func (c *Client) Stacks() stacks.Service {
	return c.stacksService
}

// Endpoints returns the endpoints service.
func (c *Client) Endpoints() endpoints.Service {
	return c.endpointsService
}

// buildURL creates the full URL for a given relative path and optional query parameters.
func (c *Client) buildURL(path string, query url.Values) string {
	cleanPath := strings.TrimLeft(path, "/")
	if !strings.HasPrefix(cleanPath, "api/") && cleanPath != "api" {
		cleanPath = "api/" + cleanPath
	}

	u := *c.baseURL
	if u.Path != "" && !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	u.Path += cleanPath

	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return u.String()
}

// prepareRequest attaches authentication headers and context to an HTTP request.
func (c *Client) prepareRequest(ctx context.Context, req *http.Request) *http.Request {
	req = req.WithContext(ctx)
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	} else if c.jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.jwtToken)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	return req
}

// Do executes an HTTP request and checks for error status codes.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = c.prepareRequest(ctx, req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil && (apiErr.Message != "" || apiErr.Details != "") {
			apiErr.StatusCode = resp.StatusCode
			return nil, &apiErr
		}

		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    strings.TrimSpace(string(body)),
		}
	}

	return resp, nil
}

// Get executes a GET request and unmarshals JSON response into target.
func (c *Client) Get(ctx context.Context, path string, query url.Values, target any) error {
	fullURL := c.buildURL(path, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if target != nil {
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// PostJSON executes a POST request with a JSON payload and unmarshals response into target.
func (c *Client) PostJSON(ctx context.Context, path string, query url.Values, payload any, target any) error {
	fullURL := c.buildURL(path, query)

	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal request payload: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bodyReader)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if target != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// PutJSON executes a PUT request with a JSON payload and unmarshals response into target.
func (c *Client) PutJSON(ctx context.Context, path string, query url.Values, payload any, target any) error {
	fullURL := c.buildURL(path, query)

	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal request payload: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fullURL, bodyReader)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if target != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// Delete executes a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, query url.Values) error {
	fullURL := c.buildURL(path, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// PostMultipart executes a multipart form data POST request.
func (c *Client) PostMultipart(ctx context.Context, path string, query url.Values, formFields map[string]string, fileFieldName string, filePath string, target any) error {
	fullURL := c.buildURL(path, query)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for k, v := range formFields {
		if err := writer.WriteField(k, v); err != nil {
			return fmt.Errorf("failed writing field %s: %w", k, err)
		}
	}

	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed opening file %s: %w", filePath, err)
		}
		defer file.Close()

		part, err := writer.CreateFormFile(fileFieldName, filepath.Base(filePath))
		if err != nil {
			return fmt.Errorf("failed creating form file: %w", err)
		}
		if _, err := io.Copy(part, file); err != nil {
			return fmt.Errorf("failed copying file contents: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if target != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed decoding response: %w", err)
		}
	}
	return nil
}
