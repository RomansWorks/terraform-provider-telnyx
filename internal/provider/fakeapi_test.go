// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// collection configures how the fake API serves one kind of object. The
// shapes follow the pinned spec; defaults stand in for server-side values.
type collection struct {
	envelope bool
	// defaults are added on create when absent, like server defaults.
	defaults map[string]any
	// created adjusts a new object, for example to add derived fields.
	created func(obj map[string]any)
	// read adjusts an object on every GET, for example to advance a status.
	read func(obj map[string]any)
}

// fakeTelnyx is an in-memory Telnyx API for unit tests: POST on a collection
// creates, and GET, PATCH, POST, PUT and DELETE act on one object.
type fakeTelnyx struct {
	t           *testing.T
	mu          sync.Mutex
	collections map[string]*collection
	objects     map[string]map[string]any
	requests    []string
	next        int
	url         string
}

func newFakeTelnyx(t *testing.T, collections map[string]*collection) *fakeTelnyx {
	t.Helper()
	f := &fakeTelnyx{t: t, collections: collections, objects: map[string]map[string]any{}}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	f.url = srv.URL
	return f
}

// seed stores an object that exists before the test, such as a phone
// number's voice settings.
func (f *fakeTelnyx) seed(path string, obj map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[path] = obj
}

func (f *fakeTelnyx) get(path string) map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.objects[path]
}

func (f *fakeTelnyx) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.objects)
}

// sent reports whether a request matching "METHOD path body" was made.
func (f *fakeTelnyx) sent(want string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.requests {
		if r == want {
			return nil
		}
	}
	return fmt.Errorf("no request %q in:\n%s", want, strings.Join(f.requests, "\n"))
}

func (f *fakeTelnyx) sentPrefix(prefix string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.requests {
		if strings.HasPrefix(r, prefix) {
			return true
		}
	}
	return false
}

func (f *fakeTelnyx) config() string {
	return fmt.Sprintf(`
provider "telnyx" {
  api_key     = "test"
  base_url    = %q
  max_retries = 0
}
`, f.url)
}

func (f *fakeTelnyx) factories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"telnyx": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func (f *fakeTelnyx) collectionOf(p string) (string, *collection) {
	best := ""
	for c := range f.collections {
		if (p == c || strings.HasPrefix(p, c+"/")) && len(c) > len(best) {
			best = c
		}
	}
	return best, f.collections[best]
}

func (f *fakeTelnyx) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Header.Get("Authorization") != "Bearer test" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	p := strings.TrimPrefix(r.URL.EscapedPath(), "/v2")
	in, raw, err := readBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f.requests = append(f.requests, strings.TrimSpace(r.Method+" "+p+" "+raw))
	name, c := f.collectionOf(p)
	if c == nil {
		http.Error(w, "no such collection: "+p, http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPost && p == name {
		f.next++
		obj := map[string]any{"id": fmt.Sprintf("%d", f.next), "created_at": "2026-10-29T00:00:00Z"}
		for k, v := range c.defaults {
			obj[k] = v
		}
		for k, v := range in {
			obj[k] = v
		}
		if c.created != nil {
			c.created(obj)
		}
		f.objects[p+"/"+obj["id"].(string)] = obj //nolint:forcetypeassert
		f.write(w, c, http.StatusOK, obj)
		return
	}
	obj, ok := f.objects[p]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"code":"10005","title":"Resource not found","detail":"The requested resource or URL could not be found."}]}`))
		return
	}
	switch r.Method {
	case http.MethodGet:
		if c.read != nil {
			c.read(obj)
		}
	case http.MethodPatch, http.MethodPost:
		for k, v := range in {
			if v == nil {
				delete(obj, k)
			} else {
				obj[k] = v
			}
		}
	case http.MethodPut:
		for k := range obj {
			if k != "id" && k != "created_at" {
				delete(obj, k)
			}
		}
		for k, v := range in {
			obj[k] = v
		}
	case http.MethodDelete:
		delete(f.objects, p)
		f.write(w, c, http.StatusOK, obj)
		return
	}
	f.write(w, c, http.StatusOK, obj)
}

// readBody decodes a JSON or multipart body. raw is the JSON as sent, or a
// summary of the multipart fields.
func readBody(r *http.Request) (map[string]any, string, error) {
	mt, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mt == "multipart/form-data" {
		mr := multipart.NewReader(r.Body, params["boundary"])
		in := map[string]any{}
		var parts []string
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, "", err
			}
			data, _ := io.ReadAll(part)
			if part.FileName() != "" {
				parts = append(parts, fmt.Sprintf("%s=file(%s,%s,%d bytes)", part.FormName(), part.FileName(), part.Header.Get("Content-Type"), len(data)))
				continue
			}
			in[part.FormName()] = string(data)
			parts = append(parts, part.FormName()+"="+string(data))
		}
		return in, "multipart " + strings.Join(parts, " "), nil
	}
	data, err := io.ReadAll(r.Body)
	if err != nil || len(data) == 0 {
		return nil, "", err
	}
	var in map[string]any
	if err := json.Unmarshal(data, &in); err != nil {
		return nil, "", err
	}
	return in, string(data), nil
}

func (f *fakeTelnyx) write(w http.ResponseWriter, c *collection, status int, obj map[string]any) {
	var out any = obj
	if c.envelope {
		out = map[string]any{"data": obj}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(out)
}
