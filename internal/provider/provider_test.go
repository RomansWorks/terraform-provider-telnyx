// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"os"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"telnyx": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck fails fast when acceptance tests run without credentials.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TELNYX_API_KEY") == "" {
		t.Fatal("TELNYX_API_KEY must be set for acceptance tests")
	}
}

func TestProviderSchemaValid(t *testing.T) {
	p := New("test")()
	var resp fwprovider.SchemaResponse
	p.Schema(context.Background(), fwprovider.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	if diags := resp.Schema.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("schema implementation: %v", diags)
	}
	if !resp.Schema.Attributes["api_key"].IsSensitive() {
		t.Error("api_key must be sensitive")
	}
}
