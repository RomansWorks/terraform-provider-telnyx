// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"
	"regexp"
	"sync"
	"testing"
)

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// Changes carry an Idempotency-Key that stays the same across retries of
// one request and differs between requests; reads carry none.
func TestIdempotencyKey(t *testing.T) {
	var mu sync.Mutex
	var keys []string
	fail := 1
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		keys = append(keys, r.Method+" "+r.Header.Get("Idempotency-Key"))
		if r.Method == http.MethodPost && fail > 0 {
			fail--
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})
	ctx := context.Background()
	for _, m := range []string{http.MethodPost, http.MethodPost, http.MethodGet, http.MethodDelete} {
		if _, err := c.Send(ctx, m, "/things", nil, "", nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(keys) != 5 {
		t.Fatalf("requests: %v", keys)
	}
	first, retry, second := keys[0][5:], keys[1][5:], keys[2][5:]
	if !uuidV4.MatchString(first) || first != retry || first == second || !uuidV4.MatchString(second) {
		t.Errorf("POST keys: %v", keys)
	}
	if keys[3] != "GET " {
		t.Errorf("GET carries a key: %q", keys[3])
	}
	if !uuidV4.MatchString(keys[4][7:]) {
		t.Errorf("DELETE key: %q", keys[4])
	}
}

// WithHeader adds headers to the requests of a context, and can replace
// the client's own.
func TestWithHeader(t *testing.T) {
	var got http.Header
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte(`{}`))
	})
	ctx := WithHeader(context.Background(), "if-match", `"3"`)
	ctx = WithHeader(ctx, "Idempotency-Key", "fixed")
	if _, err := c.Send(ctx, http.MethodPatch, "/things/1", nil, "", nil); err != nil {
		t.Fatal(err)
	}
	if got.Get("If-Match") != `"3"` || got.Get("Idempotency-Key") != "fixed" || got.Get("Authorization") != "Bearer KEY_test" {
		t.Errorf("headers: %v", got)
	}
	if _, err := c.Send(context.Background(), http.MethodGet, "/things/1", nil, "", nil); err != nil {
		t.Fatal(err)
	}
	if got.Get("If-Match") != "" {
		t.Error("a header leaked into another request")
	}
}
