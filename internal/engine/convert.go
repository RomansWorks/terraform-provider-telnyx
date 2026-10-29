// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TerraformType returns the tftypes type of an attribute.
func (a *Attr) TerraformType() tftypes.Type {
	switch a.Kind {
	case KindString, KindJSON:
		return tftypes.String
	case KindInt, KindFloat:
		return tftypes.Number
	case KindBool:
		return tftypes.Bool
	case KindList:
		return tftypes.List{ElementType: a.elemType()}
	case KindSet:
		return tftypes.Set{ElementType: a.elemType()}
	case KindObject:
		return ObjectType(a.Attrs)
	}
	panic(fmt.Sprintf("engine: unknown kind %d for %s", a.Kind, a.Name))
}

func (a *Attr) elemType() tftypes.Type {
	if a.Elem != nil {
		return a.Elem.TerraformType()
	}
	return ObjectType(a.Attrs)
}

// elem returns the element descriptor of a list or set.
func (a *Attr) elem() *Attr {
	if a.Elem != nil {
		return a.Elem
	}
	return &Attr{Name: a.Name, Kind: KindObject, Attrs: a.Attrs, InResponse: a.InResponse}
}

// ObjectType returns the tftypes object type for a set of attributes.
func ObjectType(attrs []*Attr) tftypes.Object {
	types := make(map[string]tftypes.Type, len(attrs))
	for _, a := range attrs {
		types[a.Name] = a.TerraformType()
	}
	return tftypes.Object{AttributeTypes: types}
}

// toJSON converts a Terraform value to its JSON form. ok is false for null
// and unknown values, which are never sent.
func toJSON(a *Attr, v tftypes.Value) (any, bool, error) {
	if v.IsNull() || !v.IsKnown() {
		return nil, false, nil
	}
	switch a.Kind {
	case KindString:
		var s string
		err := v.As(&s)
		return s, err == nil, err
	case KindJSON:
		var s string
		if err := v.As(&s); err != nil {
			return nil, false, err
		}
		if !json.Valid([]byte(s)) {
			return nil, false, fmt.Errorf("%s is not valid JSON", a.Name)
		}
		return json.RawMessage(s), true, nil
	case KindInt, KindFloat:
		var f big.Float
		if err := v.As(&f); err != nil {
			return nil, false, err
		}
		if a.Kind == KindInt {
			i, _ := f.Int64()
			return json.Number(strconv.FormatInt(i, 10)), true, nil
		}
		return json.Number(f.Text('g', -1)), true, nil
	case KindBool:
		var b bool
		err := v.As(&b)
		return b, err == nil, err
	case KindList, KindSet:
		var elems []tftypes.Value
		if err := v.As(&elems); err != nil {
			return nil, false, err
		}
		out := make([]any, 0, len(elems))
		ea := a.elem()
		for _, e := range elems {
			j, ok, err := toJSON(ea, e)
			if err != nil {
				return nil, false, err
			}
			if ok {
				out = append(out, j)
			}
		}
		return out, true, nil
	case KindObject:
		var fields map[string]tftypes.Value
		if err := v.As(&fields); err != nil {
			return nil, false, err
		}
		obj, err := objectToJSON(a.Attrs, fields, func(*Attr) bool { return true })
		return obj, err == nil, err
	}
	return nil, false, fmt.Errorf("engine: unknown kind for %s", a.Name)
}

// objectToJSON converts the attributes selected by include.
func objectToJSON(attrs []*Attr, fields map[string]tftypes.Value, include func(*Attr) bool) (Object, error) {
	out := Object{}
	for _, a := range attrs {
		if a.Local || !include(a) {
			continue
		}
		j, ok, err := toJSON(a, fields[a.Name])
		if err != nil {
			return nil, err
		}
		if ok {
			out[a.JSON] = j
		}
	}
	return out, nil
}

// mode selects how API values and prior values are combined.
type mode int

const (
	// modeApply builds state after create or update: values the user
	// configured (known in the plan) are kept, the API fills the rest.
	modeApply mode = iota
	// modeRead builds state from the API to detect drift, keeping prior
	// values only where the API says nothing.
	modeRead
	// modeData builds data source state: API values only.
	modeData
)

// objectFromJSON builds a Terraform object from an API object. prior holds
// the plan (modeApply) or the current state (modeRead) and may be nil.
func objectFromJSON(attrs []*Attr, obj Object, prior map[string]tftypes.Value, m mode) (tftypes.Value, error) {
	typ := ObjectType(attrs)
	vals := make(map[string]tftypes.Value, len(attrs))
	for _, a := range attrs {
		p, hasPrior := prior[a.Name]
		if !hasPrior {
			p = tftypes.NewValue(a.TerraformType(), nil)
		}
		if a.Local || !a.InResponse {
			vals[a.Name] = known(p)
			continue
		}
		raw, inObj := obj[a.JSON]
		switch {
		case m == modeApply && p.IsFullyKnown() && !p.IsNull():
			vals[a.Name] = p
			continue
		case m == modeApply && p.IsNull() && !a.Computed:
			vals[a.Name] = p
			continue
		case m != modeApply && !inObj:
			vals[a.Name] = known(p)
			continue
		}
		v, err := fromJSON(a, raw, p, m)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("%s: %w", a.Name, err)
		}
		vals[a.Name] = v
	}
	return tftypes.NewValue(typ, vals), nil
}

// fromJSON converts one API value.
func fromJSON(a *Attr, raw any, prior tftypes.Value, m mode) (tftypes.Value, error) {
	typ := a.TerraformType()
	if raw == nil {
		return tftypes.NewValue(typ, nil), nil
	}
	// Nothing is known about the value yet (for example a computed object):
	// its contents all come from the API.
	if m == modeApply && (!prior.IsKnown() || prior.IsNull()) {
		m = modeData
	}
	switch a.Kind {
	case KindString:
		switch v := raw.(type) {
		case string:
			return tftypes.NewValue(typ, v), nil
		case json.Number:
			return tftypes.NewValue(typ, v.String()), nil
		case bool:
			return tftypes.NewValue(typ, strconv.FormatBool(v)), nil
		}
		return tftypes.Value{}, fmt.Errorf("expected a string, got %T", raw)
	case KindJSON:
		data, err := canonicalJSON(raw)
		if err != nil {
			return tftypes.Value{}, err
		}
		// Keep the user's formatting when the content is the same.
		var ps string
		if prior.IsKnown() && !prior.IsNull() && prior.As(&ps) == nil && JSONEqual(ps, data) {
			return prior, nil
		}
		return tftypes.NewValue(typ, data), nil
	case KindInt, KindFloat:
		var s string
		switch v := raw.(type) {
		case json.Number:
			s = v.String()
		case float64:
			s = strconv.FormatFloat(v, 'g', -1, 64)
		case string:
			s = v
		default:
			return tftypes.Value{}, fmt.Errorf("expected a number, got %T", raw)
		}
		f, _, err := big.ParseFloat(s, 10, 512, big.ToNearestEven)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("expected a number, got %q", s)
		}
		return tftypes.NewValue(typ, f), nil
	case KindBool:
		b, ok := raw.(bool)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected a boolean, got %T", raw)
		}
		return tftypes.NewValue(typ, b), nil
	case KindList, KindSet:
		items, ok := raw.([]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected an array, got %T", raw)
		}
		var priorElems []tftypes.Value
		if prior.IsKnown() && !prior.IsNull() {
			_ = prior.As(&priorElems)
		}
		ea := a.elem()
		elems := make([]tftypes.Value, 0, len(items))
		for i, it := range items {
			var pe tftypes.Value
			if a.Kind == KindList && i < len(priorElems) {
				pe = priorElems[i]
			} else {
				pe = tftypes.NewValue(ea.TerraformType(), nil)
			}
			v, err := fromJSON(ea, it, pe, m)
			if err != nil {
				return tftypes.Value{}, fmt.Errorf("[%d]: %w", i, err)
			}
			elems = append(elems, v)
		}
		return tftypes.NewValue(typ, elems), nil
	case KindObject:
		o, ok := raw.(map[string]any)
		if !ok {
			return tftypes.Value{}, fmt.Errorf("expected an object, got %T", raw)
		}
		var pf map[string]tftypes.Value
		if prior.IsKnown() && !prior.IsNull() {
			_ = prior.As(&pf)
		}
		return objectFromJSON(a.Attrs, o, pf, m)
	}
	return tftypes.Value{}, fmt.Errorf("engine: unknown kind")
}

// known replaces unknown values, at any depth, with null: state must never
// hold unknown values.
func known(v tftypes.Value) tftypes.Value {
	if v.IsFullyKnown() {
		return v
	}
	if !v.IsKnown() {
		return tftypes.NewValue(v.Type(), nil)
	}
	switch t := v.Type().(type) {
	case tftypes.Object:
		var fields map[string]tftypes.Value
		_ = v.As(&fields)
		out := make(map[string]tftypes.Value, len(fields))
		for k, f := range fields {
			out[k] = known(f)
		}
		return tftypes.NewValue(t, out)
	case tftypes.List, tftypes.Set:
		var elems []tftypes.Value
		_ = v.As(&elems)
		out := make([]tftypes.Value, len(elems))
		for i, e := range elems {
			out[i] = known(e)
		}
		return tftypes.NewValue(t, out)
	}
	return v
}

// canonicalJSON renders a decoded JSON value compactly with sorted keys.
func canonicalJSON(v any) (string, error) {
	if rm, ok := v.(json.RawMessage); ok {
		var d any
		dec := json.NewDecoder(bytes.NewReader(rm))
		dec.UseNumber()
		if err := dec.Decode(&d); err != nil {
			return "", err
		}
		v = d
	}
	data, err := json.Marshal(v)
	return string(data), err
}

// JSONEqual reports whether two JSON documents have the same content,
// ignoring formatting and key order.
func JSONEqual(a, b string) bool {
	var va, vb any
	da := json.NewDecoder(bytes.NewReader([]byte(a)))
	da.UseNumber()
	db := json.NewDecoder(bytes.NewReader([]byte(b)))
	db.UseNumber()
	if da.Decode(&va) != nil || db.Decode(&vb) != nil {
		return false
	}
	return reflect.DeepEqual(va, vb)
}

// decodeObject decodes a JSON response body, unwrapping a "data" envelope.
func decodeObject(data []byte, envelope bool) (Object, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return Object{}, nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var obj Object
	if err := dec.Decode(&obj); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if envelope {
		inner, ok := obj["data"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("response has no data object")
		}
		return inner, nil
	}
	return obj, nil
}
