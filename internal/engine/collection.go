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
		}
		out, err := c.Send(ctx, op.Method, path, q, "", nil)
		if err != nil {
			return err
		}
		items, total, err := decodeList(out, op.Envelope)
		if err != nil {
			return err
		}
		for _, it := range items {
			if !fn(it) {
				return nil
			}
		}
		if op.Paging == PagingNone || len(items) == 0 || (total > 0 && page >= total) || (total == 0 && len(items) < size) {
			return nil
		}
	}
	return fmt.Errorf("%s %s: more than %d pages", op.Method, path, maxPages)
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

	if rf.Field == "" {
		var found Object
		err := Each(ctx, c, rf.Op, p, nil, func(it Object) bool {
			if rf.Key == "" {
				if match(it["value"]) {
					found = Object{idAttr.JSON: it["value"]}
				}
			} else if match(it[rf.Key]) {
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
		if obj, ok := el.(map[string]any); ok && match(obj[rf.Key]) {
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
