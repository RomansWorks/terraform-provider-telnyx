// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/apisync"
)

// Every operation coverage.yaml assigns to a Terraform type is built: the
// provider registers that type in the bucket's kind. Coverage alone records
// the decision, so without this check a classified operation can be left
// unbuilt and still pass.
func TestCoverageIsBuilt(t *testing.T) {
	cov, err := apisync.LoadCoverage("../../catalog/coverage.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(cov) == 0 {
		t.Fatal("coverage.yaml not found or empty")
	}
	ctx := context.Background()
	p := New("test")().(*TelnyxProvider)
	pn := "telnyx"
	types := map[string]map[string]bool{
		"resource": {}, "data_source": {}, "action": {}, "ephemeral": {},
	}
	for _, f := range p.Resources(ctx) {
		var m resource.MetadataResponse
		f().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: pn}, &m)
		types["resource"][m.TypeName] = true
	}
	for _, f := range p.DataSources(ctx) {
		var m datasource.MetadataResponse
		f().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: pn}, &m)
		types["data_source"][m.TypeName] = true
	}
	for _, f := range p.Actions(ctx) {
		var m action.MetadataResponse
		f().Metadata(ctx, action.MetadataRequest{ProviderTypeName: pn}, &m)
		types["action"][m.TypeName] = true
	}
	for _, f := range p.EphemeralResources(ctx) {
		var m ephemeral.MetadataResponse
		f().Metadata(ctx, ephemeral.MetadataRequest{ProviderTypeName: pn}, &m)
		types["ephemeral"][m.TypeName] = true
	}
	var missing []string
	for op, e := range cov {
		kinds, ok := types[e.Bucket]
		if !ok || e.Terraform == "" {
			continue
		}
		if !kinds[e.Terraform] {
			missing = append(missing, op+" -> "+e.Bucket+" "+e.Terraform)
		}
	}
	sort.Strings(missing)
	for _, m := range missing {
		t.Errorf("not built: %s", m)
	}
}
