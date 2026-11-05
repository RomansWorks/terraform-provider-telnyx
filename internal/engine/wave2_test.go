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

// TestEachRecordsPaging reads items from "records", sends page and the
// operation's size parameter, and stops once totalRecords items are read.
func TestEachRecordsPaging(t *testing.T) {
	pages := map[string]string{
		"1": `{"records":[{"brandId":"a"},{"brandId":"b"}],"page":1,"totalRecords":5}`,
		"2": `{"records":[{"brandId":"c"},{"brandId":"d"}],"page":2,"totalRecords":5}`,
		"3": `{"records":[{"brandId":"e"}],"page":3,"totalRecords":5}`,
	}
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		queries = append(queries, q.Encode())
		if q.Get("recordsPerPage") != "2" {
			t.Errorf("page size: %s", q.Encode())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pages[q.Get("page")]))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	collect := func(op *Op) []string {
		queries = nil
		var ids []string
		if err := Each(context.Background(), c, op, "/brands", nil, func(o Object) bool {
			ids = append(ids, o["brandId"].(string)) //nolint:forcetypeassert
			return true
		}); err != nil {
			t.Fatal(err)
		}
		return ids
	}
	op := &Op{Method: "GET", Path: "/brands", Envelope: true, Records: true, Paging: PagingPage, PageSize: 2, SizeParam: "recordsPerPage"}
	if ids := collect(op); strings.Join(ids, ",") != "a,b,c,d,e" || len(queries) != 3 {
		t.Errorf("ids %v, queries %v", ids, queries)
	}
	// totalRecords ends the read on a full last page.
	pages["2"] = `{"records":[{"brandId":"c"},{"brandId":"d"}],"page":2,"totalRecords":4}`
	if ids := collect(op); strings.Join(ids, ",") != "a,b,c,d" || len(queries) != 2 {
		t.Errorf("ids %v, queries %v", ids, queries)
	}
}

func TestDecodeRecords(t *testing.T) {
	items, total, err := decodeRecords([]byte(`{"records":null,"totalRecords":0}`))
	if err != nil || len(items) != 0 || total != 0 {
		t.Errorf("null records: %v %d %v", items, total, err)
	}
	if _, _, err := decodeRecords([]byte(`{"records":{"x":1}}`)); err == nil {
		t.Error("records that are not an array decoded")
	}
}

// TestGetWhole puts a response map without fixed fields into one attribute.
func TestGetWhole(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"10017":{"status":"ok"}}`))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	d := &Resource{Read: &Op{Method: "GET", Path: "/status", Whole: "values"}}
	obj, err := Get(context.Background(), c, d, Values{})
	m, ok := obj["values"].(map[string]any)
	if err != nil || !ok || m["10017"] == nil || len(obj) != 1 {
		t.Errorf("obj %v, err %v", obj, err)
	}
}
