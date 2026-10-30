// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

// Package engine implements create, read, update, delete and import once for
// every Telnyx resource. Each resource is described by a Resource descriptor,
// generated from the pinned OpenAPI spec and its override file; the engine
// builds the Terraform schema from the descriptor and converts between
// Terraform values and the JSON the API speaks.
package engine

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

// Kind is the Terraform shape of an attribute.
type Kind int

// Attribute kinds.
const (
	KindString Kind = iota
	KindInt
	KindFloat
	KindBool
	// KindJSON is a string holding JSON, compared after normalisation. Used
	// for free-form objects and unions the spec gives no stable shape for.
	KindJSON
	KindList   // list of Elem
	KindSet    // set of Elem
	KindObject // nested object with Attrs
)

// Attr describes one attribute and how it maps to the API's JSON.
type Attr struct {
	Name        string // Terraform name
	JSON        string // API field name
	Kind        Kind
	Elem        *Attr   // element of a list or set
	Attrs       []*Attr // fields of an object, or of a list/set element object
	Description string
	Enum        []string

	Required  bool
	Optional  bool
	Computed  bool
	Sensitive bool
	// ForceNew replaces the resource when the value changes.
	ForceNew bool

	// InCreate and InUpdate say whether the attribute is sent in the create
	// and update requests. InResponse says whether read returns it; when it
	// does not, the value is kept from configuration.
	InCreate   bool
	InUpdate   bool
	InResponse bool

	// Local attributes are never sent or read as JSON; hooks use them (for
	// example a file path to upload).
	Local bool

	// PathParams names the path parameters this attribute fills, such as
	// "assistant_id" for the id attribute. Operations on the same object
	// sometimes name the parameter differently.
	PathParams []string

	// Deprecated, when set, is shown to users who configure the attribute.
	Deprecated string
}

// Op is one API operation.
type Op struct {
	Method string
	Path   string
	// Envelope is true when the response wraps the object in "data".
	Envelope bool
	// Paging says how a list operation pages; PageSize is the page size to
	// ask for.
	Paging   Paging
	PageSize int
}

// Paging is a list operation's paging style.
type Paging int

// Paging styles.
const (
	// PagingNone reads the list with one request.
	PagingNone Paging = iota
	// PagingNumber sends page[number] and page[size] and stops at
	// meta.total_pages.
	PagingNumber
	// PagingFlat sends page_number and page_size and stops at
	// meta.total_pages.
	PagingFlat
)

// ReadFrom reads an object that has no item read from a collection: the
// pages of a list operation, or an array field of its parent object.
type ReadFrom struct {
	Op *Op
	// Field is the parent object's array field; empty means the list
	// response's data array.
	Field string
	// Key is the element field that holds the id; empty means the elements
	// are plain values and the value is the id.
	Key string
}

// ExtraUpdate is an operation, besides the main update, that sets some
// attributes. Clear, when set, removes them.
type ExtraUpdate struct {
	Op    *Op
	Clear *Op
	Attrs []string
}

// Resource describes one Terraform resource (and its data source).
type Resource struct {
	TypeName    string
	Description string

	Create *Op
	Read   *Op
	Update *Op
	Delete *Op

	// IDAttr is the attribute holding the object's id after create.
	IDAttr string
	// Import lists the attributes an import id fills, in order, separated
	// by "/" in the id.
	Import []string

	Attrs []*Attr
	// ExactlyOne lists groups of top-level attributes of which exactly one
	// must be set.
	ExactlyOne [][]string

	// DeleteWarning is shown on destroy when Delete is nil.
	DeleteWarning string
	// Timeouts adds a timeouts block with a create timeout.
	Timeouts bool
	// DataSource also registers a data source of the same name.
	DataSource bool
	// NoResource registers only the data source: the descriptor describes
	// read-only data.
	NoResource bool

	// ReadFrom, when set, replaces Read.
	ReadFrom *ReadFrom
	// Updates are further update operations for some attributes.
	Updates []ExtraUpdate

	Hooks Hooks
}

// Object is an API object decoded from JSON. Numbers are json.Number.
type Object = map[string]any

// Values gives hooks read access to the planned or current attribute values.
type Values map[string]tftypes.Value

// Hooks let a resource handle what the spec cannot express. All are optional.
type Hooks struct {
	// ModifyPlan can adjust planned values and request replacement. plan is
	// modified in place; state is nil on create.
	ModifyPlan func(ctx context.Context, plan, state Values) (replace []string, diags diag.Diagnostics)
	// BeforeRequest can change a create or update request body.
	BeforeRequest func(ctx context.Context, op string, body Object, plan Values) error
	// Create replaces the generic create call, for example for an upload.
	// It returns the created object as read would.
	Create func(ctx context.Context, c *client.Client, r *Resource, plan Values) (Object, error)
	// AfterRead can reshape an API object before it is mapped to attributes.
	AfterRead func(ctx context.Context, obj Object) (Object, error)
	// DeleteWarning replaces the descriptor's DeleteWarning, for example to
	// name what stays in Telnyx.
	DeleteWarning func(state Values) string
	// Wait is called after create until the object is ready.
	Wait func(ctx context.Context, c *client.Client, r *Resource, obj Object) (Object, error)
}

// List describes a data source that lists objects.
type List struct {
	TypeName    string
	Description string
	Op          *Op
	// Params are the path parameters (Required, with PathParams) and query
	// parameters (Optional; JSON is the query key, such as filter[name]).
	Params []*Attr
	// Items is the computed list attribute holding the objects.
	Items *Attr
}

// Attr returns the top-level attribute with the given name, or nil.
func (r *Resource) Attr(name string) *Attr {
	for _, a := range r.Attrs {
		if a.Name == name {
			return a
		}
	}
	return nil
}
