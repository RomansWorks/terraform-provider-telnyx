// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

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

// Every operation a resource uses is classified as that resource, reviewed,
// in catalog/coverage.yaml.
func TestCoverageNamesEveryGeneratedOperation(t *testing.T) {
	cov, err := apisync.LoadCoverage("../../../catalog/coverage.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range All() {
		for _, op := range []*engine.Op{r.Create, r.Read, r.Update, r.Delete} {
			if op == nil {
				continue
			}
			key := op.Method + " " + op.Path
			e, ok := cov[key]
			switch {
			case !ok:
				t.Errorf("%s: %s is not in coverage.yaml", r.TypeName, key)
			case e.Bucket != apisync.BucketResource || e.Terraform != r.TypeName || !e.Reviewed:
				t.Errorf("%s: coverage.yaml has %s as %s/%s reviewed=%t; want resource/%s reviewed=true",
					r.TypeName, key, e.Bucket, e.Terraform, e.Reviewed, r.TypeName)
			}
		}
	}
}
