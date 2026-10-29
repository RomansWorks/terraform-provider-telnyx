// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package apisync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func loadMini(t *testing.T) *Spec {
	t.Helper()
	s, err := LoadSpec("testdata/mini.json")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestOperationsAreKeyedByMethodAndPath(t *testing.T) {
	s := loadMini(t)
	ops := s.Operations()
	if len(ops) != 17 {
		t.Fatalf("got %d operations, want 17", len(ops))
	}
	seen := map[string]bool{}
	for _, o := range ops {
		if seen[o.Key()] {
			t.Errorf("duplicate key %s", o.Key())
		}
		seen[o.Key()] = true
	}
	// Both aliased paths share an operationId but stay distinct.
	if !seen["GET /brand/feedback/{id}"] || !seen["GET /brand_feedback/{id}"] {
		t.Error("aliased paths were not both listed")
	}
}

func TestDetectResources(t *testing.T) {
	got := map[string]Resource{}
	for _, r := range DetectResources(loadMini(t)) {
		got[r.Name] = r
	}

	a, ok := got["telnyx_ai_assistant"]
	if !ok {
		t.Fatalf("assistant not detected; got %v", reflect.ValueOf(got).MapKeys())
	}
	want := map[string]string{
		"create": "POST /ai/assistants",
		"read":   "GET /ai/assistants/{assistant_id}",
		"update": "POST /ai/assistants/{assistant_id}",
		"delete": "DELETE /ai/assistants/{assistant_id}",
		"list":   "GET /ai/assistants",
	}
	if !reflect.DeepEqual(a.Operations, want) {
		t.Errorf("assistant operations = %v, want %v", a.Operations, want)
	}
	if a.IDParam != "assistant_id" || a.Kind != KindCollection {
		t.Errorf("assistant id/kind = %s/%s", a.IDParam, a.Kind)
	}

	av := got["telnyx_ai_avatar"]
	if _, hasDelete := av.Operations["delete"]; hasDelete || av.Operations["update"] != "PATCH /ai/avatars/{avatar_id}" || !av.Beta {
		t.Errorf("avatar = %+v; want PATCH update, no delete, beta", av)
	}

	if got["telnyx_payment_auto_recharge_prefs"].Kind != KindSingleton {
		t.Errorf("auto recharge prefs should be a singleton: %+v", got["telnyx_payment_auto_recharge_prefs"])
	}
	if pn := got["telnyx_phone_number"]; pn.Kind != KindExisting || pn.Operations["delete"] != "DELETE /phone_numbers/{id}" {
		t.Errorf("phone number = %+v; want existing with delete", pn)
	}
	if len(got) != 4 {
		t.Errorf("got %d resources, want 4", len(got))
	}
}

func TestHashTracksReferencedSchemas(t *testing.T) {
	s := loadMini(t)
	before := DetectResources(s)[0].SchemaHash

	tool := asMap(asMap(asMap(s.Doc["components"])["schemas"])["Tool"])
	asMap(tool["properties"])["name"] = map[string]any{"type": "string"}
	after := DetectResources(s)[0].SchemaHash
	if before == after {
		t.Error("changing a transitively referenced schema did not change the hash")
	}
}

func TestCoverageSyncAndCheck(t *testing.T) {
	s := loadMini(t)
	res := DetectResources(s)
	cov := Coverage{
		"GET /gone": {Bucket: BucketDataSource, Reviewed: true},
		"POST /calls/{call_control_id}/actions/answer": {Bucket: BucketOutOfScope, Reviewed: true, Note: "runtime"},
	}
	r := cov.Sync(s, res)
	if len(r.Added) != 16 || !reflect.DeepEqual(r.Removed, []string{"GET /gone"}) {
		t.Errorf("sync added %d removed %v", len(r.Added), r.Removed)
	}
	if e := cov["POST /calls/{call_control_id}/actions/answer"]; !e.Reviewed {
		t.Error("sync must not touch existing entries")
	}
	if e := cov["POST /ai/assistants/{assistant_id}/clone"]; e.Bucket != BucketAction {
		t.Errorf("clone bucket = %s, want action", e.Bucket)
	}
	if e := cov["POST /ai/assistants/{assistant_id}"]; e.Bucket != BucketResource || e.Terraform != "telnyx_ai_assistant" {
		t.Errorf("assistant update entry = %+v", e)
	}

	rep := cov.Check(s)
	if !rep.OK(false) || rep.OK(true) {
		t.Errorf("check = %+v; want ok but not fully reviewed", rep)
	}

	clone := cov["POST /ai/assistants/{assistant_id}/clone"]
	clone.Reviewed = true
	cov["POST /ai/assistants/{assistant_id}/clone"] = clone
	if rep = cov.Check(s); len(rep.Invalid) != 1 {
		t.Errorf("check = %+v; want a reviewed action without a name to be invalid", rep)
	}
	clone.Terraform = "telnyx_ai_assistant_clone"
	cov["POST /ai/assistants/{assistant_id}/clone"] = clone

	delete(cov, "GET /ai/assistants")
	cov["GET /extra"] = CoverageEntry{Bucket: "bogus"}
	rep = cov.Check(s)
	if len(rep.Missing) != 1 || len(rep.Stale) != 1 || rep.OK(false) {
		t.Errorf("check = %+v; want one missing and one stale", rep)
	}
}

func TestCoverageRoundTrip(t *testing.T) {
	s := loadMini(t)
	cov := Coverage{}
	cov.Sync(s, DetectResources(s))
	p := filepath.Join(t.TempDir(), "coverage.yaml")
	if err := cov.Save(p); err != nil {
		t.Fatal(err)
	}
	back, err := LoadCoverage(p)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cov, back) {
		t.Error("coverage did not round-trip")
	}
	data, _ := os.ReadFile(p)
	if i, j := strings.Index(string(data), "GET /ai/assistants:"), strings.Index(string(data), "POST /ai/assistants:"); i < 0 || j < i {
		t.Error("entries are not grouped by path")
	}
}

func TestSplitByTagIsSelfContained(t *testing.T) {
	docs := SplitByTag(loadMini(t))
	a, ok := docs["assistants"]
	if !ok {
		t.Fatalf("no assistants document; got %v", reflect.ValueOf(docs).MapKeys())
	}
	schemas := asMap(asMap(a["components"])["schemas"])
	for _, name := range []string{"Assistant", "AssistantRequest", "Tool"} {
		if _, ok := schemas[name]; !ok {
			t.Errorf("missing referenced schema %s", name)
		}
	}
	if _, ok := schemas["Unused"]; ok {
		t.Error("unreferenced schema was included")
	}
	if _, ok := asMap(asMap(a["components"])["parameters"])["AssistantId"]; !ok {
		t.Error("path-level parameter reference was not followed")
	}
	if _, ok := docs["usage-reports-beta"]; ok {
		t.Error("unexpected tag")
	}
	if TagSlug("Usage Reports (BETA)") != "usage-reports-beta" {
		t.Errorf("TagSlug = %s", TagSlug("Usage Reports (BETA)"))
	}
}

func TestCompare(t *testing.T) {
	oldSpec := loadMini(t)
	data, _ := json.Marshal(oldSpec.Doc)
	var doc map[string]any
	_ = json.Unmarshal(data, &doc)
	newSpec := &Spec{Doc: doc}

	paths := newSpec.Paths()
	delete(paths, "/brand_feedback/{id}")
	asMap(asMap(paths["/ai/avatars/{avatar_id}"]))["delete"] = map[string]any{"operationId": "DeleteAvatar", "tags": []any{"Avatars"}}
	asMap(asMap(paths["/phone_numbers/{id}"])["patch"])["deprecated"] = true

	d := Compare(oldSpec, newSpec)
	if !reflect.DeepEqual(d.OpsAdded, []string{"DELETE /ai/avatars/{avatar_id}"}) {
		t.Errorf("added = %v", d.OpsAdded)
	}
	if !reflect.DeepEqual(d.OpsRemoved, []string{"GET /brand_feedback/{id}"}) {
		t.Errorf("removed = %v", d.OpsRemoved)
	}
	if !reflect.DeepEqual(d.OpsDeprecated, []string{"PATCH /phone_numbers/{id}"}) {
		t.Errorf("deprecated = %v", d.OpsDeprecated)
	}
	if !reflect.DeepEqual(d.ResChanged, []string{"/ai/avatars/{avatar_id}", "/phone_numbers/{id}"}) {
		t.Errorf("resources changed = %v", d.ResChanged)
	}
	md := d.Markdown()
	if !strings.Contains(md, "| Operations added | 1 |") || !strings.Contains(md, "DELETE /ai/avatars/{avatar_id}") {
		t.Errorf("markdown missing content:\n%s", md)
	}
}

func TestVerifyLock(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.json")
	lock := filepath.Join(dir, "lock.json")
	_ = os.WriteFile(spec, []byte(`{"paths":{}}`), 0o644)
	sum, _ := FileSHA256(spec)
	_ = Lock{SHA256: sum}.Save(lock)
	if err := VerifyLock(spec, lock); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(spec, []byte(`{"paths":{"/x":{}}}`), 0o644)
	if err := VerifyLock(spec, lock); err == nil {
		t.Error("hand-edited spec passed the lock check")
	}
}

func TestRawURL(t *testing.T) {
	u, err := RawURL("https://github.com/team-telnyx/openapi", "abc", "openapi/spec3.json")
	if err != nil || u != "https://raw.githubusercontent.com/team-telnyx/openapi/abc/openapi/spec3.json" {
		t.Errorf("RawURL = %s, %v", u, err)
	}
	if _, err := RawURL("https://gitlab.com/x/y", "abc", "f"); err == nil {
		t.Error("non-GitHub repo accepted")
	}
}
