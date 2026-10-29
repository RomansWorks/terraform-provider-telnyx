// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

// Package client is the HTTP core shared by every Telnyx resource and data
// source: authentication, retries with backoff, and error decoding.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the Telnyx API v2 server from the OpenAPI spec.
	DefaultBaseURL = "https://api.telnyx.com/v2"
	// DefaultMaxRetries is how many times a retryable request is re-sent.
	DefaultMaxRetries = 4
)

// Config holds the settings needed to build a Client.
type Config struct {
	APIKey     string
	BaseURL    string
	MaxRetries int
	UserAgent  string
	HTTPClient *http.Client
	// MinBackoff and MaxBackoff bound the exponential backoff between retries.
	MinBackoff time.Duration
	MaxBackoff time.Duration
}

// Client sends authenticated requests to the Telnyx API.
type Client struct {
	apiKey     string
	baseURL    *url.URL
	maxRetries int
	userAgent  string
	httpClient *http.Client
	minBackoff time.Duration
	maxBackoff time.Duration
}

// New validates cfg and returns a Client.
func New(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("telnyx: API key is required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	base, err := url.Parse(strings.TrimRight(cfg.BaseURL, "/"))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("telnyx: invalid base URL %q", cfg.BaseURL)
	}
	if cfg.MaxRetries < 0 {
		return nil, fmt.Errorf("telnyx: max retries must not be negative")
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}
	if cfg.MinBackoff <= 0 {
		cfg.MinBackoff = 500 * time.Millisecond
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 30 * time.Second
	}
	return &Client{
		apiKey:     cfg.APIKey,
		baseURL:    base,
		maxRetries: cfg.MaxRetries,
		userAgent:  cfg.UserAgent,
		httpClient: cfg.HTTPClient,
		minBackoff: cfg.MinBackoff,
		maxBackoff: cfg.MaxBackoff,
	}, nil
}

// Error is a non-2xx response from the Telnyx API.
type Error struct {
	StatusCode int
	Method     string
	Path       string
	RequestID  string
	Details    []ErrorDetail
	Body       string
}

// ErrorDetail is one entry of the Telnyx `errors` array.
type ErrorDetail struct {
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "telnyx: %s %s returned %d", e.Method, e.Path, e.StatusCode)
	for _, d := range e.Details {
		msg := d.Detail
		if msg == "" {
			msg = d.Title
		}
		if d.Code != "" {
			fmt.Fprintf(&b, "; [%s] %s", d.Code, msg)
		} else {
			fmt.Fprintf(&b, "; %s", msg)
		}
	}
	if len(e.Details) == 0 && e.Body != "" {
		fmt.Fprintf(&b, "; %s", e.Body)
	}
	if e.RequestID != "" {
		fmt.Fprintf(&b, " (request id %s)", e.RequestID)
	}
	return b.String()
}

// IsNotFound reports whether err is a 404 from the API. Resources use it to
// drop objects that were deleted outside Terraform.
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.StatusCode == http.StatusNotFound
	}
	return false
}

// Do sends a request and decodes a JSON response into out, which may be nil.
// in, when not nil, is encoded as the JSON request body.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, in, out any) error {
	var body []byte
	contentType := ""
	if in != nil {
		var err error
		if body, err = json.Marshal(in); err != nil {
			return fmt.Errorf("telnyx: encoding request body: %w", err)
		}
		contentType = "application/json"
	}
	data, err := c.Send(ctx, method, path, query, contentType, body)
	if err != nil {
		return err
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("telnyx: decoding %s %s response: %w", method, path, err)
	}
	return nil
}

// Send sends a request with a raw body of the given content type and returns
// the raw response body of a 2xx response. Non-2xx responses are returned as
// *Error. Retries follow the same rules as Do.
func (c *Client) Send(ctx context.Context, method, path string, query url.Values, contentType string, body []byte) ([]byte, error) {
	u := *c.baseURL
	u.Path = c.baseURL.Path + "/" + strings.TrimLeft(path, "/")
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("telnyx: building request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", "application/json")
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		if c.userAgent != "" {
			req.Header.Set("User-Agent", c.userAgent)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil || attempt >= c.maxRetries || !retryableMethod(method) {
				return nil, fmt.Errorf("telnyx: %s %s: %w", method, path, err)
			}
			if werr := c.wait(ctx, attempt, 0); werr != nil {
				return nil, werr
			}
			continue
		}

		data, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("telnyx: reading response: %w", readErr)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return data, nil
		}

		if attempt < c.maxRetries && retryableStatus(resp.StatusCode, method) {
			if werr := c.wait(ctx, attempt, retryAfter(resp.Header.Get("Retry-After"))); werr != nil {
				return nil, werr
			}
			continue
		}
		return nil, decodeError(resp, method, path, data)
	}
}

// retryableMethod reports whether a request can be re-sent after a transport
// error without risking a duplicate side effect.
func retryableMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	}
	return false
}

// retryableStatus reports whether a response status warrants a retry. A 429
// was never processed, so it is safe for every method; 5xx only for
// idempotent methods.
func retryableStatus(code int, method string) bool {
	if code == http.StatusTooManyRequests {
		return true
	}
	switch code {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusInternalServerError:
		return retryableMethod(method)
	}
	return false
}

func retryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if s, err := strconv.Atoi(v); err == nil && s >= 0 {
		return time.Duration(s) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func (c *Client) wait(ctx context.Context, attempt int, hint time.Duration) error {
	d := hint
	if d <= 0 {
		d = c.minBackoff << attempt
		if d > c.maxBackoff || d <= 0 {
			d = c.maxBackoff
		}
		// Full jitter keeps parallel applies from retrying in lockstep.
		d = time.Duration(rand.Int64N(int64(d)) + 1)
	} else if d > c.maxBackoff {
		d = c.maxBackoff
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func decodeError(resp *http.Response, method, path string, data []byte) error {
	e := &Error{
		StatusCode: resp.StatusCode,
		Method:     method,
		Path:       path,
		RequestID:  firstHeader(resp.Header, "X-Request-Id", "X-Telnyx-Request-Id"),
	}
	var payload struct {
		Errors []ErrorDetail `json:"errors"`
	}
	if json.Unmarshal(data, &payload) == nil && len(payload.Errors) > 0 {
		e.Details = payload.Errors
	} else {
		e.Body = truncate(strings.TrimSpace(string(data)), 512)
	}
	return e
}

func firstHeader(h http.Header, names ...string) string {
	for _, n := range names {
		if v := h.Get(n); v != "" {
			return v
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
