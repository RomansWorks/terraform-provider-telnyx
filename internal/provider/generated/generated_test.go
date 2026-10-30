// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package generated

import (
	"context"
	"testing"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
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
		if !r.NoResource {
			if diags := r.ResourceSchema(ctx).ValidateImplementation(ctx); diags.HasError() {
				t.Errorf("%s resource schema: %v", r.TypeName, diags)
			}
		}
		if r.DataSource {
			if diags := r.DataSourceSchema().ValidateImplementation(ctx); diags.HasError() {
				t.Errorf("%s data source schema: %v", r.TypeName, diags)
			}
		}
		if !r.NoResource && r.Attr(r.IDAttr) == nil {
			t.Errorf("%s has no id attribute %s", r.TypeName, r.IDAttr)
		}
	}
	for _, l := range Lists() {
		if seen[l.TypeName] {
			t.Errorf("%s is generated twice", l.TypeName)
		}
		seen[l.TypeName] = true
		if diags := l.Schema().ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s list schema: %v", l.TypeName, diags)
		}
	}
}

// Every operation a resource uses is classified as that resource, reviewed,
// in catalog/coverage.yaml.
func TestCoverageNamesEveryGeneratedOperation(t *testing.T) {
	cov, err := apisync.LoadCoverage("../../../catalog/coverage.yaml")
	if err != nil {
		t.Fatal(err)
	}
	check := func(name, key, bucket string) {
		t.Helper()
		e, ok := cov[key]
		switch {
		case !ok:
			t.Errorf("%s: %s is not in coverage.yaml", name, key)
		case e.Bucket != bucket || e.Terraform != name || !e.Reviewed:
			t.Errorf("%s: coverage.yaml has %s as %s/%s reviewed=%t; want %s/%s reviewed=true",
				name, key, e.Bucket, e.Terraform, e.Reviewed, bucket, name)
		}
	}
	for _, r := range All() {
		if r.NoResource {
			check(r.TypeName, r.Read.Method+" "+r.Read.Path, apisync.BucketDataSource)
			continue
		}
		ops := []*engine.Op{r.Create, r.Read, r.Update, r.Delete}
		for _, u := range r.Updates {
			ops = append(ops, u.Op, u.Clear)
		}
		for _, op := range ops {
			if op != nil {
				check(r.TypeName, op.Method+" "+op.Path, apisync.BucketResource)
			}
		}
		// A read_from list may also back a list data source, so its bucket
		// is not fixed; it only has to be in coverage.yaml.
		if r.ReadFrom != nil {
			if _, ok := cov[r.ReadFrom.Op.Method+" "+r.ReadFrom.Op.Path]; !ok {
				t.Errorf("%s: read_from %s is not in coverage.yaml", r.TypeName, r.ReadFrom.Op.Path)
			}
		}
	}
	for _, l := range Lists() {
		check(l.TypeName, l.Op.Method+" "+l.Op.Path, apisync.BucketDataSource)
	}
}
