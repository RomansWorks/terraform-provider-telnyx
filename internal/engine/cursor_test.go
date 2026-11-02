// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// TestEachCursorPaging follows both cursor shapes Telnyx uses until the
// cursor is empty.
func TestEachCursorPaging(t *testing.T) {
	pages := map[string]string{
		"":   `{"data":[{"id":"a"},{"id":"b"}],"meta":{"cursors":{"after":"c1"}}}`,
		"c1": `{"data":[{"id":"c"},{"id":"d"}],"meta":{"end_cursor":"c2"}}`,
		"c2": `{"data":[{"id":"e"}],"meta":{"cursors":{"after":""}}}`,
	}
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		queries = append(queries, q.Encode())
		if q.Get("page[size]") != "2" {
			t.Errorf("page size: %s", q.Encode())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pages[q.Get("page[after]")]))
	}))
	t.Cleanup(srv.Close)
	c, err := client.New(client.Config{APIKey: "test", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	op := &Op{Method: "GET", Path: "/things", Envelope: true, Paging: PagingCursor, PageSize: 2}
	var ids []string
	err = Each(context.Background(), c, op, "/things", nil, func(o Object) bool {
		ids = append(ids, o["id"].(string)) //nolint:forcetypeassert
		return true
	})
	if err != nil || strings.Join(ids, ",") != "a,b,c,d,e" || len(queries) != 3 {
		t.Errorf("ids %v, err %v, queries %v", ids, err, queries)
	}
	if strings.Contains(queries[0], "after") {
		t.Errorf("first page sent a cursor: %s", queries[0])
	}
}

func TestNextCursor(t *testing.T) {
	for in, want := range map[string]string{
		`{"meta":{"cursors":{"after":"x"}}}`: "x",
		`{"meta":{"end_cursor":"y"}}`:        "y",
		`{"meta":{"end_cursor":null}}`:       "",
		`{"meta":{"page_number":1}}`:         "",
		`not json`:                           "",
	} {
		if got := nextCursor([]byte(in)); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

func TestWithFixedID(t *testing.T) {
	r := &Resource{FixedID: "settings"}
	in := Object{"channels": json.Number("3")}
	out := r.withFixedID(in)
	if out["id"] != "settings" || out["channels"] != json.Number("3") || in["id"] != nil {
		t.Errorf("out %v, in %v", out, in)
	}
	if got := (&Resource{}).withFixedID(in); got["id"] != nil {
		t.Errorf("no fixed id: %v", got)
	}
}
