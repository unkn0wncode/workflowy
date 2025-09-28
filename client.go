// Package workflowy / client.go defines the API Client for Workflowy.
package workflowy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// BaseURL is the default base URL for the Workflowy API.
const BaseURL = "https://workflowy.com/api/v1"

// Client is a minimal HTTP client for the Workflowy API.
//
// It authenticates using an API key provided at construction time. Requests
// are sent with an Authorization header using the Bearer scheme.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient constructs a client using the default base URL, default HTTP client, and given API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     strings.TrimSpace(apiKey),
		baseURL:    BaseURL,
		httpClient: DefaultHTTPClient(),
	}
}

// DefaultHTTPClient creates a default HTTP client used for new API Clients.
func DefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Second}
}

// SetBaseURL sets the client's base URL. Empty resets to the default.
func (c *Client) SetBaseURL(baseURL string) {
	trimmedBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedBase == "" {
		trimmedBase = BaseURL
	}
	c.baseURL = trimmedBase
}

// SetHTTPClient overrides the underlying HTTP client.
func (c *Client) SetHTTPClient(hc *http.Client) {
	if hc == nil {
		return
	}
	c.httpClient = hc
}

// APIError represents a non-2xx response from the API.
type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message,omitempty"`
	ErrorText  string `json:"error,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return fmt.Sprintf("workflowy api error (%d): %s", e.StatusCode, e.Message)
	}
	if e.ErrorText != "" {
		return fmt.Sprintf("workflowy api error (%d): %s", e.StatusCode, e.ErrorText)
	}
	return fmt.Sprintf("workflowy api error (%d)", e.StatusCode)
}

// joinURL joins the base URL with the given path.
func (c *Client) joinURL(path string) string {
	if path == "" || path == "/" {
		return c.baseURL
	}
	return c.baseURL + "/" + strings.TrimLeft(path, "/")
}

// newRequest creates a new request with the given context, method, path, and body.
func (c *Client) newRequest(ctx context.Context, method string, path string, body any) (*http.Request, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	var (
		rbody   io.Reader
		rawBody []byte
	)
	if body != nil {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
		rawBody = buf.Bytes()
		rbody = bytes.NewReader(rawBody)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.joinURL(path), rbody)
	if err != nil {
		return nil, err
	}
	if rawBody != nil {
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(rawBody)), nil
		}
		req.ContentLength = int64(len(rawBody))
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "workflowy-go/0.1 (+https://workflowy.com)")
	return req, nil
}

// do executes a request and handles 429 Too Many Requests errors by retrying.
func (c *Client) do(req *http.Request, v any) error {
	const max429Retries = 3
	for attempt := 0; ; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusTooManyRequests { // 429
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
			resp.Body.Close()
			if retryAfter <= 0 || attempt >= max429Retries {
				return &APIError{StatusCode: resp.StatusCode, Message: "too many requests"}
			}
			select {
			case <-req.Context().Done():
				return req.Context().Err()
			case <-time.After(retryAfter):
			}
			if req.GetBody != nil && req.Body != nil {
				if b, gerr := req.GetBody(); gerr == nil {
					req.Body = b
				}
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			apiErr := &APIError{StatusCode: resp.StatusCode}
			_ = json.NewDecoder(resp.Body).Decode(apiErr)
			resp.Body.Close()
			return apiErr
		}
		if v == nil {
			resp.Body.Close()
			return nil
		}
		err = json.NewDecoder(resp.Body).Decode(v)
		resp.Body.Close()
		return err
	}
}

// parseRetryAfter parses a Retry-After header value which may be either a number of seconds
// or an HTTP-date. Returns a duration to wait from now.
func parseRetryAfter(h string, now time.Time) time.Duration {
	if h == "" {
		return 0
	}
	if secs, err := strconv.ParseInt(strings.TrimSpace(h), 10, 64); err == nil {
		if secs < 0 {
			secs = 0
		}
		return time.Duration(secs) * time.Second
	}
	if retryAt, err := http.ParseTime(h); err == nil {
		interval := retryAt.Sub(now)
		if interval < 0 {
			return 0
		}
		return interval
	}
	return 0
}
