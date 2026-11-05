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

// Later pages of a page[number] list carry meta.snapshot from the first,
// and has_more false ends the list.
func TestEachSnapshotPaging(t *testing.T) {
	pages := map[string]string{
		"1": `{"data":[{"id":"a"},{"id":"b"}],"meta":{"page_number":1,"page_size":2,"has_more":true,"snapshot":"s1"}}`,
		"2": `{"data":[{"id":"c"},{"id":"d"}],"meta":{"page_number":2,"page_size":2,"has_more":false,"snapshot":"s1"}}`,
	}
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		queries = append(queries, q.Get("page[number]")+":"+q.Get("page[snapshot]"))
		_, _ = w.Write([]byte(pages[q.Get("page[number]")]))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	op := &Op{Method: "GET", Path: "/keys", Envelope: true, Paging: PagingNumber, PageSize: 2}
	var ids []string
	if err := Each(context.Background(), c, op, "/keys", nil, func(o Object) bool {
		ids = append(ids, scalarString(o["id"]))
		return true
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(ids, ",") != "a,b,c,d" || strings.Join(queries, "|") != "1:|2:s1" {
		t.Errorf("ids %v, queries %v", ids, queries)
	}
}

// A short results page is not the end while count says more remain.
func TestEachResultsShortPage(t *testing.T) {
	pages := map[string]string{
		"1": `{"count":3,"results":[{"id":1}]}`,
		"2": `{"count":3,"results":[{"id":2},{"id":3}]}`,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(pages[r.URL.Query().Get("page")]))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	op := &Op{Method: "GET", Path: "/agents", Envelope: true, Results: true, Paging: PagingPage, PageSize: 2, SizeParam: "page_size"}
	n := 0
	if err := Each(context.Background(), c, op, "/agents", nil, func(Object) bool { n++; return true }); err != nil || n != 3 {
		t.Errorf("read %d items: %v", n, err)
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
