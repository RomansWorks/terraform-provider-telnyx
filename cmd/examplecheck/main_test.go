// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

func TestCompleteAddsMissingObjects(t *testing.T) {
	l := &library{root: t.TempDir(), schemas: map[string]block{}}
	srcs := map[string][]byte{"main.tf": []byte(`resource "telnyx_a" "x" {
  b_id = telnyx_b.y.id
  name = var.name
  cert = file("${path.module}/cert.pem")
}
data "telnyx_c" "z" {
  id = data.telnyx_d.w.id
}
`)}
	l.complete(srcs)
	if got := string(srcs["stub_telnyx_b_y.tf"]); !strings.HasPrefix(got, `resource "telnyx_b" "y" {`) {
		t.Errorf("stub for telnyx_b.y = %q", got)
	}
	if _, ok := srcs["stub_telnyx_d_w.tf"]; ok {
		t.Error("data source reference was stubbed as a resource")
	}
	if got := string(srcs["stub_var_name.tf"]); got != "variable \"name\" {}\n" {
		t.Errorf("variable stub = %q", got)
	}
	if m := fileRe.FindAllStringSubmatch(string(srcs["main.tf"]), -1); len(m) != 1 || m[0][1] != "cert.pem" {
		t.Errorf("file references = %v", m)
	}
}

func TestExtractRenamesFirstBlock(t *testing.T) {
	src := "resource \"telnyx_b\" \"other\" {\n  name = \"n\"\n}\n"
	if got := extract(src, `resource "telnyx_b" "y"`); got != "" {
		t.Errorf("extract found %q", got)
	}
	if got := extract(src, `resource "telnyx_b" `); !strings.Contains(got, `name = "n"`) {
		t.Errorf("extract = %q", got)
	}
}
