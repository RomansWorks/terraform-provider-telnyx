// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// TestEachResultsPaging reads items from "results", sends page and
// page_size, and stops once count items are read.
func TestEachResultsPaging(t *testing.T) {
	pages := map[string]string{
		"1": `{"count":3,"next":"https://x/?page=2","previous":null,"results":[{"id":1},{"id":2}]}`,
		"2": `{"count":3,"next":null,"previous":"https://x/?page=1","results":[{"id":3}]}`,
	}
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		queries = append(queries, q.Encode())
		_, _ = w.Write([]byte(pages[q.Get("page")]))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	op := &Op{Method: "GET", Path: "/agents", Envelope: true, Results: true, Paging: PagingPage, PageSize: 2, SizeParam: "page_size"}
	var ids []string
	if err := Each(context.Background(), c, op, "/agents", nil, func(o Object) bool {
		ids = append(ids, scalarString(o["id"]))
		return true
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(ids, ",") != "1,2,3" || strings.Join(queries, "|") != "page=1&page_size=2|page=2&page_size=2" {
		t.Errorf("ids %v, queries %v", ids, queries)
	}
}

func TestDecodeItems(t *testing.T) {
	items, total, err := decodeItems([]byte(`{"results":null,"count":0}`), "results", "count")
	if err != nil || len(items) != 0 || total != 0 {
		t.Errorf("null results: %v %d %v", items, total, err)
	}
	items, total, err = decodeItems([]byte(`{"results":["a"],"count":7}`), "results", "count")
	if err != nil || len(items) != 1 || items[0]["value"] != "a" || total != 7 {
		t.Errorf("plain results: %v %d %v", items, total, err)
	}
}

// Fields only the create response carries are kept for the state built
// from the read after create; fields the read has win.
func TestWithCreated(t *testing.T) {
	attrs := []*Attr{
		{Name: "token", JSON: "token", Kind: KindString, CreateOnly: true},
		{Name: "name", JSON: "name", Kind: KindString},
		{Name: "missing", JSON: "missing", Kind: KindString, CreateOnly: true},
	}
	read := Object{"name": "b"}
	out := withCreated(attrs, read, Object{"token": "secret", "name": "a"})
	if out["token"] != "secret" || out["name"] != "b" || len(out) != 2 {
		t.Errorf("merged: %v", out)
	}
	if _, ok := read["token"]; ok {
		t.Error("the read object was modified")
	}
	if got := withCreated(attrs, read, Object{}); len(got) != 1 {
		t.Errorf("nothing to add: %v", got)
	}
}
