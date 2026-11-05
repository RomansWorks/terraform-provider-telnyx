// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// DefaultPageSize is the page size asked for when the spec gives none.
const DefaultPageSize = 250

// maxPages bounds a list read, so a server that never reports its last
// page cannot loop forever.
const maxPages = 10000

// Each pages through a list operation and calls fn with every item, in API
// order, until fn returns false or the last page has been read.
func Each(ctx context.Context, c *client.Client, op *Op, path string, query url.Values, fn func(Object) bool) error {
	size := op.PageSize
	if size <= 0 {
		size = DefaultPageSize
	}
	cursor := ""
	read := 0
	for page := 1; page <= maxPages; page++ {
		q := url.Values{}
		for k, v := range query {
			q[k] = v
		}
		switch op.Paging {
		case PagingNumber:
			q.Set("page[number]", strconv.Itoa(page))
			q.Set("page[size]", strconv.Itoa(size))
		case PagingFlat:
			q.Set("page_number", strconv.Itoa(page))
			q.Set("page_size", strconv.Itoa(size))
		case PagingPage:
			q.Set("page", strconv.Itoa(page))
			q.Set(op.SizeParam, strconv.Itoa(size))
		case PagingCursor:
			q.Set("page[size]", strconv.Itoa(size))
			if cursor != "" {
				q.Set("page[after]", cursor)
			}
		}
		out, err := c.Send(ctx, op.Method, path, q, "", nil)
		if err != nil {
			return err
		}
		var items []Object
		var total int
		if op.Records || op.Results {
			field, count := "records", "totalRecords"
			if op.Results {
				field, count = "results", "count"
			}
			var records int
			items, records, err = decodeItems(out, field, count)
			read += len(items)
			if err == nil && records > 0 && read >= records {
				total = page
			}
		} else {
			items, total, err = decodeList(out, op.Envelope)
		}
		if err != nil {
			return err
		}
		for _, it := range items {
			if !fn(it) {
				return nil
			}
		}
		if op.Paging == PagingCursor {
			next := nextCursor(out)
			if next == "" || next == cursor || len(items) == 0 {
				return nil
			}
			cursor = next
			continue
		}
		if op.Paging == PagingNone || len(items) == 0 || (total > 0 && page >= total) || (total == 0 && len(items) < size) {
			return nil
		}
	}
	return fmt.Errorf("%s %s: more than %d pages", op.Method, path, maxPages)
}

// nextCursor returns the cursor of the next page: meta.cursors.after or
// meta.end_cursor, whichever the response carries. Empty means last page.
func nextCursor(data []byte) string {
	var body struct {
		Meta struct {
			Cursors struct {
				After *string `json:"after"`
			} `json:"cursors"`
			EndCursor *string `json:"end_cursor"`
		} `json:"meta"`
	}
	if json.Unmarshal(data, &body) != nil {
		return ""
	}
	if body.Meta.Cursors.After != nil {
		return *body.Meta.Cursors.After
	}
	if body.Meta.EndCursor != nil {
		return *body.Meta.EndCursor
	}
	return ""
}

// decodeList decodes a list response: its data array and meta.total_pages
// (0 when the response does not say). Without an envelope the response is
// the bare array.
func decodeList(data []byte, envelope bool) ([]Object, int, error) {
	var body struct {
		Data []json.RawMessage `json:"data"`
		Meta struct {
			TotalPages json.Number `json:"total_pages"`
		} `json:"meta"`
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var err error
	if envelope {
		err = dec.Decode(&body)
	} else {
		err = dec.Decode(&body.Data)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("decoding list response: %w", err)
	}
	items := make([]Object, 0, len(body.Data))
	for _, raw := range body.Data {
		var v any
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		if err := d.Decode(&v); err != nil {
			return nil, 0, fmt.Errorf("decoding list item: %w", err)
		}
		obj, ok := v.(map[string]any)
		if !ok {
			// Plain values become an object with one "value" field.
			obj = Object{"value": v}
		}
		items = append(items, obj)
	}
	total, _ := strconv.Atoi(body.Meta.TotalPages.String())
	return items, total, nil
}

// decodeRecords decodes a list response that holds its items in "records"
// and their count in "totalRecords" (0 when the response does not say).
func decodeRecords(data []byte) ([]Object, int, error) {
	return decodeItems(data, "records", "totalRecords")
}

// decodeItems decodes a list response that holds its items in field
// ("records" or "results") and their count in count ("totalRecords" or
// "count"; 0 when the response does not say).
func decodeItems(data []byte, field, count string) ([]Object, int, error) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, 0, fmt.Errorf("decoding list response: %w", err)
	}
	var total json.Number
	if raw, ok := body[count]; ok {
		_ = json.Unmarshal(raw, &total)
	}
	records := body[field]
	if len(records) == 0 || string(records) == "null" {
		return nil, 0, nil
	}
	items, _, err := decodeList(records, false)
	n, _ := strconv.Atoi(total.String())
	return items, n, err
}

// errNotInCollection is the not-found error of a read from a collection,
// shaped like a 404 so callers treat both the same.
func errNotInCollection(op *Op, path, id string) error {
	return &client.Error{Method: op.Method, Path: path, StatusCode: http.StatusNotFound,
		Body: fmt.Sprintf("%s is not in the collection", id)}
}

// readFrom finds the object in its collection.
func readFrom(ctx context.Context, c *client.Client, d *Resource, vals Values) (Object, error) {
	rf := d.ReadFrom
	idAttr := d.Attr(d.IDAttr)
	var id string
	if idAttr == nil || vals[d.IDAttr].As(&id) != nil || id == "" {
		return nil, fmt.Errorf("%s is not known", d.IDAttr)
	}
	p, err := fillPath(d.Attrs, rf.Op.Path, vals)
	if err != nil {
		return nil, err
	}
	match := func(v any) bool { return scalarString(v) == id }
	// Further fields that identify the object must match too.
	others := map[string]string{}
	for _, n := range rf.Match {
		a := d.Attr(n)
		var s string
		if a == nil || vals[n].As(&s) != nil || s == "" {
			return nil, fmt.Errorf("%s is not known", n)
		}
		others[a.JSON] = s
	}
	matchObj := func(obj map[string]any) bool {
		if !match(obj[rf.Key]) {
			return false
		}
		for k, s := range others {
			if scalarString(obj[k]) != s {
				return false
			}
		}
		return true
	}

	if rf.Field == "" {
		var found Object
		err := Each(ctx, c, rf.Op, p, nil, func(it Object) bool {
			if rf.Key == "" {
				if match(it["value"]) {
					found = Object{idAttr.JSON: it["value"]}
				}
			} else if matchObj(it) {
				found = it
			}
			return found == nil
		})
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, errNotInCollection(rf.Op, p, id)
		}
		return found, nil
	}

	out, err := c.Send(ctx, rf.Op.Method, p, nil, "", nil)
	if err != nil {
		return nil, err
	}
	parent, err := Decode(out, rf.Op.Envelope)
	if err != nil {
		return nil, err
	}
	list, _ := parent[rf.Field].([]any)
	for _, el := range list {
		if rf.Key == "" {
			if match(el) {
				return Object{idAttr.JSON: el}, nil
			}
			continue
		}
		if obj, ok := el.(map[string]any); ok && matchObj(obj) {
			return obj, nil
		}
	}
	return nil, errNotInCollection(rf.Op, p, id)
}

// scalarString renders a JSON scalar the way it appears in a path or id.
func scalarString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case bool:
		return strconv.FormatBool(x)
	case nil:
		return ""
	}
	return fmt.Sprint(v)
}
