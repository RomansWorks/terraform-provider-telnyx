// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package apisync

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Coverage buckets. Every operation in the pinned spec has exactly one.
const (
	BucketResource   = "resource"     // part of a managed resource's lifecycle
	BucketDataSource = "data_source"  // read-only lookup exposed as a data source
	BucketAction     = "action"       // one-shot operation exposed as a Terraform action
	BucketEphemeral  = "ephemeral"    // short-lived value such as a token, exposed as an ephemeral resource
	BucketOutOfScope = "out_of_scope" // runtime traffic or reporting; not infrastructure
)

// ValidBuckets lists the allowed bucket values.
var ValidBuckets = map[string]bool{
	BucketResource: true, BucketDataSource: true, BucketAction: true,
	BucketEphemeral: true, BucketOutOfScope: true,
}

// CoverageEntry records how one operation is handled by the provider.
type CoverageEntry struct {
	Bucket string `yaml:"bucket"`
	// Terraform is the resource, data source, action or ephemeral resource
	// type name that implements this operation.
	Terraform string `yaml:"terraform,omitempty"`
	// Reviewed is false for entries filled in by the heuristic and true once a
	// person or agent has confirmed the bucket during an OpenSpec change.
	Reviewed bool   `yaml:"reviewed"`
	Note     string `yaml:"note,omitempty"`
}

// Coverage maps "METHOD /path" to its entry.
type Coverage map[string]CoverageEntry

// runtimeTags are product areas whose operations act on live traffic or
// produce reports rather than describe infrastructure.
var runtimeTags = map[string]string{
	"Anthropic Messages":             "runtime: model inference",
	"Audio":                          "runtime: media processing",
	"Call Commands":                  "runtime: live call control",
	"Chat":                           "runtime: model inference",
	"Conference Commands":            "runtime: live call control",
	"Embeddings":                     "runtime: model inference",
	"Messages":                       "runtime: message sending",
	"OpenAI Chat":                    "runtime: model inference",
	"OpenAI Embeddings":              "runtime: model inference",
	"Programmable Fax Commands":      "runtime: fax sending",
	"Queue Commands":                 "runtime: live call control",
	"Research":                       "runtime: model inference",
	"Speech To Text over WebSockets": "runtime: streaming",
	"TeXML REST Commands":            "runtime: live call control",
	"Text To Speech Commands":        "runtime: media processing",
	"Web Search":                     "runtime: model inference",
	"Audit Logs":                     "reporting",
	"CDR Reports":                    "reporting",
	"CDR Usage Reports":              "reporting",
	"Detail Records":                 "reporting",
	"MDR Detail Reports":             "reporting",
	"MDR Detailed Reports":           "reporting",
	"MDR Usage Reports":              "reporting",
	"Reporting":                      "reporting",
	"Reports":                        "reporting",
	"Speech to Text Batch Reports":   "reporting",
	"Speech to text Usage Reports":   "reporting",
	"Telco Data Usage Reports":       "reporting",
	"Usage Reports (BETA)":           "reporting",
	"WDR Detail Reports":             "reporting",
	"Voice SDK Stats":                "reporting",
	"SETI Observability":             "reporting",
	"Bucket Usage":                   "reporting",
	"Debugging":                      "reporting",
	"Session Analysis":               "reporting",
	"Conversation Histories":         "runtime: conversation data",
	"Meeting Session Data":           "runtime: meeting data",
	"Meeting Session Artifacts":      "runtime: meeting data",
	"Meeting Session Actions":        "runtime: live meeting control",
	"Speech To Text Capabilities":    "reference data",
	"Machine Payments":               "runtime: payments",
	"x402 Payment Transactions":      "runtime: payments",
	"Stored Payment Transactions":    "runtime: payments",
	"Whatsapp messaging":             "runtime: message sending",
	"Email Messages":                 "runtime: message sending",
	"Email Events":                   "reporting",
	"Email Validations":              "runtime: validation",
	"Number Lookup":                  "runtime: lookup",
	"Verify":                         "runtime: verification flow",
	"Verification Requests":          "runtime: verification flow",
	"Rooms Client Tokens":            "runtime: client tokens",
	"Presigned Object URLs":          "runtime: object access",
	"Bot Signup":                     "runtime: account signup",
	"Whatsapp Signup":                "runtime: account signup",
	"OAuth Protocol":                 "runtime: OAuth flow",
	"OAuth Discovery":                "reference data",
	"Customer Service Record":        "runtime: carrier lookup",
	"Inventory Level":                "reference data",
	"Country Coverage":               "reference data",
	"Coverage":                       "reference data",
	"Mobile Network Operators":       "reference data",
	"Regions":                        "reference data",
	"Wireless Regions":               "reference data",
	"Reference Data":                 "reference data",
	"Pricing":                        "reference data",
	"Requirement Types":              "reference data",
	"Noise Suppression Engines":      "reference data",
	"Enum":                           "reference data",
}

// Classify proposes a bucket for an operation. resourceOf maps operation
// keys to the catalog resource that owns them.
func Classify(op Operation, resourceOf map[string]Resource) CoverageEntry {
	if r, ok := resourceOf[op.Key()]; ok {
		return CoverageEntry{Bucket: BucketResource, Terraform: r.Name, Note: "auto: " + r.Kind + " lifecycle"}
	}
	if why, ok := runtimeTags[op.Tag()]; ok {
		if why == "reference data" && op.Method == "get" {
			return CoverageEntry{Bucket: BucketDataSource, Note: "auto: reference data"}
		}
		return CoverageEntry{Bucket: BucketOutOfScope, Note: "auto: " + why}
	}
	if op.Method == "get" {
		return CoverageEntry{Bucket: BucketDataSource, Note: "auto: read-only"}
	}
	return CoverageEntry{Bucket: BucketAction, Note: "auto: non-lifecycle mutation"}
}

// ResourceIndex maps every operation key used by a resource to that resource.
func ResourceIndex(resources []Resource) map[string]Resource {
	idx := map[string]Resource{}
	for _, r := range resources {
		for _, k := range sortedKeys(r.Operations) {
			key := r.Operations[k]
			if _, taken := idx[key]; !taken {
				idx[key] = r
			}
		}
	}
	return idx
}

// LoadCoverage reads a coverage file. A missing file is an empty coverage.
func LoadCoverage(path string) (Coverage, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Coverage{}, nil
	}
	if err != nil {
		return nil, err
	}
	c := Coverage{}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return c, nil
}

// Save writes the coverage file with a header explaining how to edit it.
func (c Coverage) Save(path string) error {
	// Order by path, then method, so all operations on a path sit together.
	keys := sortedKeys(c)
	sort.SliceStable(keys, func(i, j int) bool {
		mi, pi, _ := strings.Cut(keys[i], " ")
		mj, pj, _ := strings.Cut(keys[j], " ")
		if pi != pj {
			return pi < pj
		}
		return methodRank(mi) < methodRank(mj)
	})
	root := &yaml.Node{Kind: yaml.MappingNode}
	for _, k := range keys {
		var v yaml.Node
		if err := v.Encode(c[k]); err != nil {
			return err
		}
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: k}, &v)
	}
	data, err := yaml.Marshal(root)
	if err != nil {
		return err
	}
	header := "# Coverage of every operation in upstream/telnyx/spec3.json.\n" +
		"# Generated entries have reviewed: false. Confirm or change the bucket in an\n" +
		"# OpenSpec change, then set reviewed: true. `make catalog` adds new operations\n" +
		"# and removes deleted ones; it never changes an existing entry.\n" +
		"# Buckets: resource, data_source, action, ephemeral, out_of_scope.\n"
	return os.WriteFile(path, append([]byte(header), data...), 0o644)
}

func methodRank(m string) int {
	for i, x := range Methods {
		if strings.EqualFold(x, m) {
			return i
		}
	}
	return len(Methods)
}

// SyncResult reports what Sync changed.
type SyncResult struct {
	Added   []string
	Removed []string
}

// Sync adds entries for new operations and removes entries for operations
// no longer in the spec. Existing entries are never modified.
func (c Coverage) Sync(s *Spec, resources []Resource) SyncResult {
	idx := ResourceIndex(resources)
	live := map[string]bool{}
	var res SyncResult
	for _, op := range s.Operations() {
		k := op.Key()
		live[k] = true
		if _, ok := c[k]; !ok {
			c[k] = Classify(op, idx)
			res.Added = append(res.Added, k)
		}
	}
	for _, k := range sortedKeys(c) {
		if !live[k] {
			delete(c, k)
			res.Removed = append(res.Removed, k)
		}
	}
	return res
}

// CheckReport is the result of Check.
type CheckReport struct {
	Missing    []string
	Stale      []string
	Invalid    []string
	Unreviewed int
	ByBucket   map[string]int
}

// OK reports whether the coverage is complete and valid.
func (r CheckReport) OK(requireReviewed bool) bool {
	ok := len(r.Missing) == 0 && len(r.Stale) == 0 && len(r.Invalid) == 0
	if requireReviewed {
		ok = ok && r.Unreviewed == 0
	}
	return ok
}

func (r CheckReport) String() string {
	var b strings.Builder
	buckets := sortedKeys(r.ByBucket)
	fmt.Fprintf(&b, "coverage:")
	for _, k := range buckets {
		fmt.Fprintf(&b, " %s=%d", k, r.ByBucket[k])
	}
	fmt.Fprintf(&b, " unreviewed=%d\n", r.Unreviewed)
	for _, k := range r.Missing {
		fmt.Fprintf(&b, "missing: %s is in the spec but not in coverage.yaml (run make catalog)\n", k)
	}
	for _, k := range r.Stale {
		fmt.Fprintf(&b, "stale: %s is in coverage.yaml but not in the spec (run make catalog)\n", k)
	}
	for _, k := range r.Invalid {
		fmt.Fprintf(&b, "invalid: %s\n", k)
	}
	return b.String()
}

// Check verifies that every operation is classified exactly once with a
// valid bucket, that resource and ephemeral entries name a Terraform type,
// that reviewed data source and action entries name one too, and that
// out_of_scope entries give a reason and name none.
func (c Coverage) Check(s *Spec) CheckReport {
	r := CheckReport{ByBucket: map[string]int{}}
	live := map[string]bool{}
	for _, op := range s.Operations() {
		k := op.Key()
		live[k] = true
		e, ok := c[k]
		if !ok {
			r.Missing = append(r.Missing, k)
			continue
		}
		r.ByBucket[e.Bucket]++
		if !e.Reviewed {
			r.Unreviewed++
		}
		switch {
		case !ValidBuckets[e.Bucket]:
			r.Invalid = append(r.Invalid, fmt.Sprintf("%s has unknown bucket %q", k, e.Bucket))
		case e.Bucket != BucketOutOfScope && e.Terraform == "" &&
			(e.Reviewed || e.Bucket == BucketResource || e.Bucket == BucketEphemeral):
			r.Invalid = append(r.Invalid, fmt.Sprintf("%s is bucket %s but names no terraform type", k, e.Bucket))
		case e.Bucket == BucketOutOfScope && e.Note == "":
			r.Invalid = append(r.Invalid, fmt.Sprintf("%s is out_of_scope without a note saying why", k))
		case e.Bucket == BucketOutOfScope && e.Terraform != "":
			r.Invalid = append(r.Invalid, fmt.Sprintf("%s is out_of_scope but names terraform type %s", k, e.Terraform))
		}
	}
	for _, k := range sortedKeys(c) {
		if !live[k] {
			r.Stale = append(r.Stale, k)
		}
	}
	sort.Strings(r.Invalid)
	return r
}
