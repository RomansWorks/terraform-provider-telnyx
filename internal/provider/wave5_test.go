// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// stubAPI serves one test's hand-written handler, for resources whose API
// the descriptor-driven fake cannot play: raw bodies, uploads, headers and
// membership endpoints. It records every request as "METHOD path?query body".
type stubAPI struct {
	mu       sync.Mutex
	url      string
	requests []string
}

// stubRequest is one request as a handler sees it.
type stubRequest struct {
	*http.Request
	path string // without /v2
	body string // raw, or a summary of multipart fields
}

func newStub(t *testing.T, h func(r stubRequest) (int, string)) *stubAPI {
	t.Helper()
	s := &stubAPI{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		body, err := stubBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		req := stubRequest{Request: r, path: strings.TrimPrefix(r.URL.Path, "/v2"), body: body}
		line := r.Method + " " + req.path
		if r.URL.RawQuery != "" {
			line += "?" + r.URL.RawQuery
		}
		s.requests = append(s.requests, strings.TrimSpace(line+" "+body))
		status, out := h(req)
		if strings.HasPrefix(out, "{") || strings.HasPrefix(out, "[") {
			w.Header().Set("Content-Type", "application/json")
		} else if out != "" {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(out))
	}))
	t.Cleanup(srv.Close)
	s.url = srv.URL + "/v2"
	return s
}

func stubBody(r *http.Request) (string, error) {
	mt, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mt != "multipart/form-data" {
		data, err := io.ReadAll(r.Body)
		return string(data), err
	}
	mr := multipart.NewReader(r.Body, params["boundary"])
	var parts []string
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		data, _ := io.ReadAll(part)
		if part.FileName() != "" {
			parts = append(parts, fmt.Sprintf("%s=file(%s,%s,%q)", part.FormName(), part.FileName(), part.Header.Get("Content-Type"), data))
			continue
		}
		parts = append(parts, part.FormName()+"="+string(data))
	}
	slices.Sort(parts)
	return "multipart " + strings.Join(parts, " "), nil
}

func (s *stubAPI) config() string {
	return fmt.Sprintf(`
provider "telnyx" {
  api_key     = "test"
  base_url    = %q
  max_retries = 0
}
`, strings.TrimSuffix(s.url, "/v2"))
}

func (s *stubAPI) sent(want string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slices.Contains(s.requests, want) {
		return nil
	}
	return fmt.Errorf("no request %q in:\n%s", want, strings.Join(s.requests, "\n"))
}

func stubFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"telnyx": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func jsonOf(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

const notFound = `{"errors":[{"code":"10005","title":"Resource not found"}]}`

// KV keys store their value as the raw request body and read it back raw.
func TestStorageKVKey(t *testing.T) {
	values := map[string]string{}
	api := newStub(t, func(r stubRequest) (int, string) {
		switch r.Method {
		case http.MethodPut:
			if r.Header.Get("Content-Type") != "application/octet-stream" {
				return http.StatusUnsupportedMediaType, ""
			}
			_, existed := values[r.path]
			values[r.path] = r.body
			if existed {
				return http.StatusOK, ""
			}
			return http.StatusCreated, ""
		case http.MethodGet:
			v, ok := values[r.path]
			if !ok {
				return http.StatusNotFound, notFound
			}
			return http.StatusOK, v
		case http.MethodDelete:
			delete(values, r.path)
			return http.StatusOK, ""
		}
		return http.StatusMethodNotAllowed, ""
	})
	cfg := func(value string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_storage_kv_key" "x" {
  namespace_id = "ns1"
  key          = "greeting"
  value        = %q
  ttl_secs     = 3600
}
data "telnyx_storage_kv_key" "x" {
  namespace_id = telnyx_storage_kv_key.x.namespace_id
  key          = telnyx_storage_kv_key.x.key
}`, value)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		CheckDestroy: func(*terraform.State) error {
			if len(values) != 0 {
				return fmt.Errorf("keys left: %v", values)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: cfg("hello\nworld"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_storage_kv_key.x", "key", "greeting"),
					resource.TestCheckResourceAttr("data.telnyx_storage_kv_key.x", "value", "hello\nworld"),
					func(*terraform.State) error {
						return api.sent("PUT /storage/kvs/ns1/keys/greeting?ttl_secs=3600 hello\nworld")
					},
				),
			},
			{
				// The value changes in place: the same PUT.
				Config: cfg("bye"),
				Check: func(*terraform.State) error {
					return api.sent("PUT /storage/kvs/ns1/keys/greeting?ttl_secs=3600 bye")
				},
			},
			{
				// A change outside Terraform shows as drift.
				PreConfig:          func() { values["/storage/kvs/ns1/keys/greeting"] = "changed" },
				Config:             cfg("bye"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{Config: cfg("bye")},
			{
				ResourceName: "telnyx_storage_kv_key.x", ImportState: true, ImportStateId: "ns1/greeting",
				ImportStateVerify: true, ImportStateVerifyIgnore: []string{"ttl_secs"}, ImportStateVerifyIdentifierAttribute: "key",
			},
		},
	})
}

// A tag is read back from the assistant's tags.
func TestAIAssistantTag(t *testing.T) {
	var tags []string
	api := newStub(t, func(r stubRequest) (int, string) {
		switch {
		case r.Method == http.MethodPost && r.path == "/ai/assistants/a1/tags":
			var in struct{ Tag string }
			_ = json.Unmarshal([]byte(r.body), &in)
			tags = append(tags, in.Tag)
		case r.Method == http.MethodDelete && strings.HasPrefix(r.path, "/ai/assistants/a1/tags/"):
			tags = slices.DeleteFunc(tags, func(s string) bool { return s == strings.TrimPrefix(r.path, "/ai/assistants/a1/tags/") })
		case r.Method == http.MethodGet && r.path == "/ai/assistants/a1":
			return http.StatusOK, jsonOf(map[string]any{"id": "a1", "name": "Support", "tags": tags})
		default:
			return http.StatusNotFound, notFound
		}
		return http.StatusOK, jsonOf(map[string]any{"tags": tags})
	})
	cfg := func(tag string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_ai_assistant_tag" "x" {
  assistant_id = "a1"
  tag          = %q
}`, tag)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		CheckDestroy: func(*terraform.State) error {
			if len(tags) != 0 {
				return fmt.Errorf("tags left: %v", tags)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: cfg("vip"),
				Check: func(*terraform.State) error {
					return api.sent(`POST /ai/assistants/a1/tags {"tag":"vip"}`)
				},
			},
			{
				// Removed outside Terraform: created again.
				PreConfig: func() { tags = nil },
				Config:    cfg("vip"),
				Check: func(*terraform.State) error {
					if !slices.Equal(tags, []string{"vip"}) {
						return fmt.Errorf("tags: %v", tags)
					}
					return nil
				},
			},
			{
				// A new tag replaces the old one.
				Config: cfg("gold"),
				Check: func(*terraform.State) error {
					if !slices.Equal(tags, []string{"gold"}) {
						return fmt.Errorf("tags: %v", tags)
					}
					return api.sent("DELETE /ai/assistants/a1/tags/vip")
				},
			},
			{
				ResourceName: "telnyx_ai_assistant_tag.x", ImportState: true, ImportStateId: "a1/gold",
				ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "tag",
			},
		},
	})
}

// An insight template's assignment is read back from its group.
func TestAIConversationInsightGroupAssignment(t *testing.T) {
	var insights []map[string]any
	api := newStub(t, func(r stubRequest) (int, string) {
		const group = "/ai/conversations/insight-groups/g1"
		switch {
		case r.Method == http.MethodPost && r.path == group+"/insights/i1/assign":
			insights = append(insights, map[string]any{"id": "i1", "name": "Sentiment", "instructions": "Rate it"})
			return http.StatusOK, ""
		case r.Method == http.MethodDelete && r.path == group+"/insights/i1/unassign":
			insights = nil
			return http.StatusOK, ""
		case r.Method == http.MethodGet && r.path == group:
			return http.StatusOK, jsonOf(map[string]any{"data": map[string]any{"id": "g1", "name": "Default", "insights": insights}})
		}
		return http.StatusNotFound, notFound
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		CheckDestroy: func(*terraform.State) error {
			if len(insights) != 0 {
				return fmt.Errorf("still assigned: %v", insights)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: api.config() + `
resource "telnyx_ai_conversation_insight_group_assignment" "x" {
  group_id   = "g1"
  insight_id = "i1"
}`,
				Check: resource.TestCheckResourceAttr("telnyx_ai_conversation_insight_group_assignment.x", "insight_id", "i1"),
			},
			{
				ResourceName: "telnyx_ai_conversation_insight_group_assignment.x", ImportState: true,
				ImportStateId: "g1/i1", ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "insight_id",
			},
		},
	})
}

// Gateway token keys need an Idempotency-Key on every change and the
// version in If-Match on update and delete. The key's token is only in the
// create response and stays in state.
func TestAIGatewayTokenKey(t *testing.T) {
	var key map[string]any
	api := newStub(t, func(r stubRequest) (int, string) {
		if r.Method != http.MethodGet && r.Header.Get("Idempotency-Key") == "" {
			return http.StatusBadRequest, `{"errors":[{"detail":"Idempotency-Key is required"}]}`
		}
		ifMatch := func() bool { return r.Header.Get("If-Match") == fmt.Sprintf(`"%v"`, key["version"]) }
		switch {
		case r.Method == http.MethodPost && r.path == "/llm_token_gateway/token_keys":
			_ = json.Unmarshal([]byte(r.body), &key)
			key["id"], key["version"], key["blocked"] = "k1", 1, false
			created := map[string]any{"token": "sk-secret"}
			for k, v := range key {
				created[k] = v
			}
			return http.StatusCreated, jsonOf(map[string]any{"data": created})
		case key == nil:
			return http.StatusNotFound, notFound
		case r.Method == http.MethodGet:
			return http.StatusOK, jsonOf(map[string]any{"data": key})
		case !ifMatch():
			return http.StatusPreconditionFailed, `{"errors":[{"detail":"stale version"}]}`
		case r.Method == http.MethodPatch:
			var in map[string]any
			_ = json.Unmarshal([]byte(r.body), &in)
			for k, v := range in {
				key[k] = v
			}
			key["version"] = key["version"].(int) + 1 //nolint:forcetypeassert
			return http.StatusOK, jsonOf(map[string]any{"data": key})
		case r.Method == http.MethodDelete:
			key = nil
			return http.StatusNoContent, ""
		}
		return http.StatusNotFound, notFound
	})
	cfg := func(name string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_ai_gateway_token_key" "x" {
  name           = %q
  token_group_id = "g1"
}`, name)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		CheckDestroy: func(*terraform.State) error {
			if key != nil {
				return fmt.Errorf("key left: %v", key)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: cfg("ci"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_ai_gateway_token_key.x", "token", "sk-secret"),
					resource.TestCheckResourceAttr("telnyx_ai_gateway_token_key.x", "version", "1"),
				),
			},
			{
				Config: cfg("ci-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_ai_gateway_token_key.x", "token", "sk-secret"),
					resource.TestCheckResourceAttr("telnyx_ai_gateway_token_key.x", "version", "2"),
					func(*terraform.State) error {
						return api.sent(`PATCH /llm_token_gateway/token_keys/k1 {"name":"ci-2"}`)
					},
				),
			},
			{
				ResourceName: "telnyx_ai_gateway_token_key.x", ImportState: true, ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

// Voice clones are made from a voice design with JSON or from a recording
// with a multipart upload, and read from the list of clones.
func TestVoiceClone(t *testing.T) {
	clones := map[string]map[string]any{}
	next := 0
	api := newStub(t, func(r stubRequest) (int, string) {
		create := func(fields map[string]any) (int, string) {
			next++
			id := fmt.Sprintf("c%d", next)
			obj := map[string]any{"id": id, "record_type": "voice_clone", "status": "active", "provider": "telnyx"}
			for k, v := range fields {
				obj[k] = v
			}
			clones[id] = obj
			return http.StatusCreated, jsonOf(map[string]any{"data": obj})
		}
		switch {
		case r.Method == http.MethodPost && r.path == "/voice_clones":
			var in map[string]any
			_ = json.Unmarshal([]byte(r.body), &in)
			in["source_voice_design_id"] = in["voice_design_id"]
			delete(in, "voice_design_id")
			return create(in)
		case r.Method == http.MethodPost && r.path == "/voice_clones/from_upload":
			fields := map[string]any{}
			for _, kv := range strings.Fields(strings.TrimPrefix(r.body, "multipart ")) {
				if k, v, ok := strings.Cut(kv, "="); ok && !strings.HasPrefix(v, "file(") {
					fields[k] = v
				}
			}
			return create(fields)
		case r.Method == http.MethodGet && r.path == "/voice_clones":
			var data []map[string]any
			for _, id := range slices.Sorted(func(yield func(string) bool) {
				for k := range clones {
					if !yield(k) {
						return
					}
				}
			}) {
				data = append(data, clones[id])
			}
			return http.StatusOK, jsonOf(map[string]any{"data": data, "meta": map[string]any{"total_pages": 1}})
		case r.Method == http.MethodPatch:
			obj, ok := clones[strings.TrimPrefix(r.path, "/voice_clones/")]
			if !ok {
				return http.StatusNotFound, notFound
			}
			_ = json.Unmarshal([]byte(r.body), &obj)
			return http.StatusOK, jsonOf(map[string]any{"data": obj})
		case r.Method == http.MethodDelete:
			delete(clones, strings.TrimPrefix(r.path, "/voice_clones/"))
			return http.StatusNoContent, ""
		}
		return http.StatusNotFound, notFound
	})
	audio := filepath.Join(t.TempDir(), "me.wav")
	if err := os.WriteFile(audio, []byte("RIFF"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := func(name string) string {
		return api.config() + fmt.Sprintf(`
resource "telnyx_voice_clone" "designed" {
  name            = %q
  voice_design_id = "d1"
  language        = "en"
  gender          = "female"
}
resource "telnyx_voice_clone" "recorded" {
  name       = "Me"
  audio_path = %q
  language   = "en"
  gender     = "male"
  ref_text   = "Hello there"
}`, name, audio)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		CheckDestroy: func(*terraform.State) error {
			if len(clones) != 0 {
				return fmt.Errorf("clones left: %v", clones)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: cfg("Narrator"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("telnyx_voice_clone.designed", "source_voice_design_id", "d1"),
					resource.TestCheckResourceAttr("telnyx_voice_clone.recorded", "audio_sha256", sha256Hex("RIFF")),
					func(*terraform.State) error {
						return api.sent(`POST /voice_clones {"gender":"female","language":"en","name":"Narrator","voice_design_id":"d1"}`)
					},
					func(*terraform.State) error {
						return api.sent(`POST /voice_clones/from_upload multipart audio_file=file(me.wav,audio/wav,"RIFF") gender=male language=en name=Me ref_text=Hello there`)
					},
				),
			},
			{
				Config: cfg("Narrator 2"),
				Check: func(*terraform.State) error {
					return api.sent(`PATCH /voice_clones/c1 {"name":"Narrator 2"}`)
				},
			},
		},
	})
}

// A bucket's certificate and key are uploaded as two PEM files.
func TestStorageBucketSSLCertificate(t *testing.T) {
	var cert map[string]any
	api := newStub(t, func(r stubRequest) (int, string) {
		if r.path != "/storage/buckets/assets/ssl_certificate" {
			return http.StatusNotFound, notFound
		}
		switch r.Method {
		case http.MethodPut:
			cert = map[string]any{"id": "cert-1", "issued_to": map[string]any{"common_name": "cdn.example.com"}}
		case http.MethodDelete:
			if cert == nil {
				return http.StatusNotFound, notFound
			}
			out := cert
			cert = nil
			return http.StatusOK, jsonOf(map[string]any{"data": out})
		}
		if cert == nil {
			return http.StatusNotFound, notFound
		}
		return http.StatusOK, jsonOf(map[string]any{"data": cert})
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: stubFactories(),
		CheckDestroy: func(*terraform.State) error {
			if cert != nil {
				return fmt.Errorf("certificate left")
			}
			return nil
		},
		Steps: []resource.TestStep{{
			Config: api.config() + `
resource "telnyx_storage_bucket_ssl_certificate" "x" {
  bucket_name     = "assets"
  certificate_pem = "CERT"
  private_key_pem = "KEY"
}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("telnyx_storage_bucket_ssl_certificate.x", "issued_to.common_name", "cdn.example.com"),
				func(*terraform.State) error {
					return api.sent(`PUT /storage/buckets/assets/ssl_certificate multipart certificate=file(certificate.pem,application/x-pem-file,"CERT") private_key=file(private_key.pem,application/x-pem-file,"KEY")`)
				},
			),
		}},
	})
}
