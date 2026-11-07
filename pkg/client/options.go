package client

import (
	"crypto/tls"
	"net/http"
	"time"
)

// Option is a functional option for configuring a Client.
type Option func(*Client) error

// WithBaseURL sets the base URL for the Portainer API (e.g. https://localhost:9443).
func WithBaseURL(rawURL string) Option {
	return func(c *Client) error {
		return c.setBaseURL(rawURL)
	}
}

// WithAPIKey sets the Portainer API key for authentication (via X-API-Key header).
func WithAPIKey(key string) Option {
	return func(c *Client) error {
		c.apiKey = key
		return nil
	}
}

// WithJWT sets the Bearer JWT token for authentication.
func WithJWT(token string) Option {
	return func(c *Client) error {
		c.jwtToken = token
		return nil
	}
}

// WithInsecureTLS enables or disables skipping TLS certificate verification.
func WithInsecureTLS(insecure bool) Option {
	return func(c *Client) error {
		c.insecureTLS = insecure
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig == nil {
				transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecure} // #nosec G402
			} else {
				transport.TLSClientConfig.InsecureSkipVerify = insecure
			}
		}
		return nil
	}
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		c.httpClient.Timeout = timeout
		return nil
	}
}

// WithHTTPClient overrides the internal http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient != nil {
			c.httpClient = httpClient
		}
		return nil
	}
}
