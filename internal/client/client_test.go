// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := New(Config{
		APIKey:     "KEY_test",
		BaseURL:    srv.URL + "/v2",
		MaxRetries: 3,
		UserAgent:  "terraform-provider-telnyx/test",
		MinBackoff: time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewValidation(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Error("expected error for missing API key")
	}
	if _, err := New(Config{APIKey: "k", BaseURL: "not a url"}); err == nil {
		t.Error("expected error for invalid base URL")
	}
	if _, err := New(Config{APIKey: "k", MaxRetries: -1}); err == nil {
		t.Error("expected error for negative retries")
	}
	c, err := New(Config{APIKey: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL.String() != DefaultBaseURL {
		t.Errorf("base URL = %s, want %s", c.baseURL, DefaultBaseURL)
	}
}

func TestDoSendsAuthAndDecodes(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer KEY_test" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "terraform-provider-telnyx/test" {
			t.Errorf("User-Agent = %q", got)
		}
		if r.URL.Path != "/v2/ai/assistants" || r.URL.Query().Get("page[size]") != "2" {
			t.Errorf("unexpected URL %s", r.URL)
		}
		var in map[string]string
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in["name"] != "support" {
			t.Errorf("body name = %q", in["name"])
		}
		_, _ = w.Write([]byte(`{"id":"a1"}`))
	})
	var out struct{ ID string }
	q := map[string][]string{"page[size]": {"2"}}
	if err := c.Do(context.Background(), http.MethodPost, "/ai/assistants", q, map[string]string{"name": "support"}, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != "a1" {
		t.Errorf("id = %q", out.ID)
	}
}

func TestDoRetriesRateLimitForPost(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	if err := c.Do(context.Background(), http.MethodPost, "/x", nil, map[string]int{"a": 1}, nil); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 3 {
		t.Errorf("calls = %d, want 3", calls.Load())
	}
}

func TestDoDoesNotRetryServerErrorForPost(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	})
	err := c.Do(context.Background(), http.MethodPost, "/x", nil, nil, nil)
	if err == nil || calls.Load() != 1 {
		t.Fatalf("err = %v, calls = %d; want error after 1 call", err, calls.Load())
	}
}

func TestDoRetriesServerErrorForGetThenGivesUp(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	err := c.Do(context.Background(), http.MethodGet, "/x", nil, nil, nil)
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("err = %v", err)
	}
	if calls.Load() != 4 {
		t.Errorf("calls = %d, want 4 (1 + 3 retries)", calls.Load())
	}
}

func TestErrorDecoding(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "req-123")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"code":"10005","title":"Resource not found","detail":"Assistant not found"}]}`))
	})
	err := c.Do(context.Background(), http.MethodGet, "/ai/assistants/nope", nil, nil, nil)
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound(%v) = false", err)
	}
	msg := err.Error()
	for _, want := range []string{"404", "[10005] Assistant not found", "req-123"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q does not contain %q", msg, want)
		}
	}
}

func TestDoHonoursContextCancel(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	c.minBackoff, c.maxBackoff = time.Second, time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := c.Do(ctx, http.MethodGet, "/x", nil, nil, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want deadline exceeded", err)
	}
}
