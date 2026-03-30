package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "https://api.bitbucket.org/2.0"
	userAgent      = "bb-cli/dev"
)

// Client wraps http.Client with Bitbucket-specific helpers.
type Client struct {
	http    *http.Client
	baseURL string
}

// New creates an API Client using the provided http.Client (which should have
// auth middleware already configured via NewAuthTransport or NewTokenTransport).
func New(httpClient *http.Client, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{http: httpClient, baseURL: baseURL}
}

// Get performs a GET request to path (relative to baseURL) and decodes the JSON
// response body into dest.
func (c *Client) Get(path string, dest any) error {
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dest)
}

// GetRaw performs a GET and returns the raw response body bytes.
// The caller is responsible for closing the body.
func (c *Client) GetRaw(path string, extraHeaders map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	return c.http.Do(req)
}

// Post performs a POST request with a JSON body and decodes the response into dest.
func (c *Client) Post(path string, body any, dest any) error {
	resp, err := c.do("POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if dest != nil {
		return json.NewDecoder(resp.Body).Decode(dest)
	}
	return nil
}

// Put performs a PUT request with a JSON body and decodes the response into dest.
func (c *Client) Put(path string, body any, dest any) error {
	resp, err := c.do("PUT", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if dest != nil {
		return json.NewDecoder(resp.Body).Decode(dest)
	}
	return nil
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	resp, err := c.do("DELETE", path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		pr, pw := io.Pipe()
		go func() {
			enc := json.NewEncoder(pw)
			pw.CloseWithError(enc.Encode(body))
		}()
		bodyReader = pr
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	if err := checkStatus(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

// checkStatus returns an error for non-2xx responses.
func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Try to extract Bitbucket's error message format
	var apiErr struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error.Message)
	}

	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
}

// retryTransport wraps an http.RoundTripper with exponential backoff on 429 responses.
type retryTransport struct {
	base     http.RoundTripper
	maxRetry int
}

func NewRetryTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &retryTransport{base: base, maxRetry: 3}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	backoff := 2 * time.Second

	for attempt := 0; attempt <= t.maxRetry; attempt++ {
		resp, err = t.base.RoundTrip(req)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests {
			return resp, err
		}
		if attempt == t.maxRetry {
			break
		}
		resp.Body.Close()
		time.Sleep(backoff)
		backoff *= 2
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
	}
	return resp, err
}

// TokenTransport authenticates requests using HTTP Basic auth with a Bitbucket API token.
// This is the fallback for CI/headless environments (replaces deprecated app passwords).
type TokenTransport struct {
	Username string
	Token    string
	Base     http.RoundTripper
}

func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.SetBasicAuth(t.Username, t.Token)
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}
