// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package generated

import (
	"context"
	"testing"
)

// Every generated descriptor must yield valid resource and data source
// schemas.
func TestSchemasAreValid(t *testing.T) {
	ctx := context.Background()
	seen := map[string]bool{}
	for _, r := range All() {
		if seen[r.TypeName] {
			t.Errorf("%s is generated twice", r.TypeName)
		}
		seen[r.TypeName] = true
		if diags := r.ResourceSchema(ctx).ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s resource schema: %v", r.TypeName, diags)
		}
		if r.DataSource {
			if diags := r.DataSourceSchema().ValidateImplementation(ctx); diags.HasError() {
				t.Errorf("%s data source schema: %v", r.TypeName, diags)
			}
		}
		if r.Attr(r.IDAttr) == nil {
			t.Errorf("%s has no id attribute %s", r.TypeName, r.IDAttr)
		}
	}
}
