// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	rtest "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
)

func providers(c *client.Client, p *testProvider) map[string]func() (tfprotov6.ProviderServer, error) {
	p.c = c
	return map[string]func() (tfprotov6.ProviderServer, error){
		"telnyx": providerserver.NewProtocol6WithError(p),
		"echo":   echoprovider.NewProviderServer(),
	}
}

// vault has a write-only password.
func vault() *Resource {
	return &Resource{
		TypeName: "telnyx_vault",
		Create:   &Op{Method: "POST", Path: "/vaults", Envelope: true},
		Read:     &Op{Method: "GET", Path: "/vaults/{id}", Envelope: true},
		Update:   &Op{Method: "PATCH", Path: "/vaults/{id}", Envelope: true},
		Delete:   &Op{Method: "DELETE", Path: "/vaults/{id}"},
		IDAttr:   "id",
		Import:   []string{"id"},
		Attrs: []*Attr{
			{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true, PathParams: []string{"id"}},
			{Name: "name", JSON: "name", Kind: KindString, Required: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: "password_wo", JSON: "password", Kind: KindString, Optional: true, Sensitive: true, WriteOnly: true,
				Version: "password_wo_version", InCreate: true, InUpdate: true},
			{Name: "password_wo_version", Kind: KindInt, Optional: true, Local: true},
		},
	}
}

// number is kept in Telnyx on destroy unless release_on_destroy is true.
func number() *Resource {
	return &Resource{
		TypeName:         "telnyx_number",
		Create:           &Op{Method: "POST", Path: "/numbers", Envelope: true},
		Read:             &Op{Method: "GET", Path: "/numbers/{id}", Envelope: true},
		Update:           &Op{Method: "PATCH", Path: "/numbers/{id}", Envelope: true},
		Delete:           &Op{Method: "DELETE", Path: "/numbers/{id}"},
		ReleaseOnDestroy: true,
		IDAttr:           "id",
		Import:           []string{"id"},
		Attrs: []*Attr{
			{Name: "id", JSON: "id", Kind: KindString, Computed: true, InResponse: true, PathParams: []string{"id"}},
			{Name: "name", JSON: "name", Kind: KindString, Required: true, InCreate: true, InUpdate: true, InResponse: true},
			{Name: ReleaseAttr, Kind: KindBool, Optional: true, Local: true},
		},
	}
}

func setBlocklist() *Action {
	return &Action{
		TypeName: "telnyx_box_set_blocklist",
		Op:       &Op{Method: "POST", Path: "/boxes/{box_id}/actions/set_blocklist", Envelope: true},
		Attrs: []*Attr{
			{Name: "box_id", Kind: KindString, Required: true, Local: true, PathParams: []string{"box_id"}},
			{Name: "blocklist_id", JSON: "blocklist_id", Kind: KindString, Required: true, InCreate: true},
		},
	}
}

func boxName() *Ephemeral {
	return &Ephemeral{
		TypeName: "telnyx_box_name",
		Open:     &Op{Method: "GET", Path: "/boxes/{box_id}", Envelope: true},
		Attrs: []*Attr{
			{Name: "box_id", Kind: KindString, Required: true, Local: true, PathParams: []string{"box_id"}},
			{Name: "name", JSON: "name", Kind: KindString, Computed: true, InResponse: true},
		},
	}
}

func TestWriteOnly(t *testing.T) {
	api, c := newFakeAPI(t)
	cfg := func(pw string, version int) string {
		return fmt.Sprintf(`resource "telnyx_vault" "v" {
  name                = "v"
  password_wo         = %q
  password_wo_version = %d
}`, pw, version)
	}
	rtest.UnitTest(t, rtest.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)},
		ProtoV6ProviderFactories: factories(c, vault()),
		Steps: []rtest.TestStep{
			{
				Config: cfg("first", 1),
				Check: rtest.ComposeAggregateTestCheckFunc(
					rtest.TestCheckNoResourceAttr("telnyx_vault.v", "password_wo"),
					rtest.TestCheckResourceAttr("telnyx_vault.v", "password_wo_version", "1"),
					func(*terraform.State) error {
						if !api.has(`POST /vaults {"name":"v","password":"first"}`) {
							return fmt.Errorf("create did not send the password: %v", api.log())
						}
						return nil
					},
				),
			},
			{
				// A new value alone is invisible to Terraform.
				Config: cfg("second", 1),
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: cfg("second", 2),
				ConfigPlanChecks: rtest.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("telnyx_vault.v", plancheck.ResourceActionUpdate)},
				},
				Check: func(*terraform.State) error {
					if !api.has(`PATCH /vaults/vaults-1 {"password":"second"}`) {
						return fmt.Errorf("update did not send only the password: %v", api.log())
					}
					return nil
				},
			},
		},
	})
	for _, r := range api.log() {
		if strings.HasPrefix(r, "PATCH") && strings.Contains(r, "first") {
			t.Errorf("old password sent again: %s", r)
		}
	}
}

func TestReleaseOnDestroy(t *testing.T) {
	t.Run("kept by default", func(t *testing.T) {
		api, c := newFakeAPI(t)
		rtest.UnitTest(t, rtest.TestCase{
			ProtoV6ProviderFactories: factories(c, number()),
			CheckDestroy: func(*terraform.State) error {
				if api.has("DELETE") || api.count() != 1 {
					return fmt.Errorf("the number must stay in Telnyx: %v", api.log())
				}
				return nil
			},
			Steps: []rtest.TestStep{{Config: `resource "telnyx_number" "n" { name = "n" }`}},
		})
	})
	t.Run("released when asked", func(t *testing.T) {
		api, c := newFakeAPI(t)
		rtest.UnitTest(t, rtest.TestCase{
			ProtoV6ProviderFactories: factories(c, number()),
			CheckDestroy: func(*terraform.State) error {
				if !api.has("DELETE /numbers/numbers-1") || api.count() != 0 {
					return fmt.Errorf("the number must be released: %v", api.log())
				}
				return nil
			},
			Steps: []rtest.TestStep{{Config: `resource "telnyx_number" "n" {
  name               = "n"
  release_on_destroy = true
}`}},
		})
	})
}

func TestAction(t *testing.T) {
	api, c := newCollectionAPI(t)
	trigger := func(box string) string {
		return fmt.Sprintf(`
action "telnyx_box_set_blocklist" "b" {
  config {
    box_id       = %q
    blocklist_id = "bl-9"
  }
}
resource "terraform_data" "t" {
  input = %q
  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.telnyx_box_set_blocklist.b]
    }
  }
}`, box, box)
	}
	rtest.UnitTest(t, rtest.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		ProtoV6ProviderFactories: providers(c, &testProvider{actions: []*Action{setBlocklist()}}),
		Steps: []rtest.TestStep{
			{
				Config: trigger("box-1"),
				Check: func(*terraform.State) error {
					if api.sent(`POST /boxes/box-1/actions/set_blocklist? {"blocklist_id":"bl-9"}`) != 1 {
						return fmt.Errorf("action not sent once: %v", api.log())
					}
					return nil
				},
			},
			{
				Config:      trigger("nope"),
				ExpectError: regexpMust(`Unable to run telnyx_box_set_blocklist`),
			},
		},
	})
}

func TestEphemeral(t *testing.T) {
	_, c := newCollectionAPI(t)
	rtest.UnitTest(t, rtest.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_10_0)},
		ProtoV6ProviderFactories: providers(c, &testProvider{ephemerals: []*Ephemeral{boxName()}}),
		Steps: []rtest.TestStep{
			{
				Config: `
ephemeral "telnyx_box_name" "n" {
  box_id = "box-1"
}
provider "echo" {
  data = ephemeral.telnyx_box_name.n.name
}
resource "echo" "n" {}`,
				Check: rtest.TestCheckResourceAttr("echo.n", "data", "b"),
			},
		},
	})
}

func TestEphemeralError(t *testing.T) {
	_, c := newCollectionAPI(t)
	rtest.UnitTest(t, rtest.TestCase{
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_10_0)},
		ProtoV6ProviderFactories: providers(c, &testProvider{ephemerals: []*Ephemeral{boxName()}}),
		Steps: []rtest.TestStep{
			{
				Config: `
ephemeral "telnyx_box_name" "n" {
  box_id = "nope"
}
provider "echo" {
  data = ephemeral.telnyx_box_name.n.name
}
resource "echo" "n" {}`,
				ExpectError: regexpMust(`Unable to open telnyx_box_name`),
			},
		},
	})
}

func TestExtraSchemasAreValid(t *testing.T) {
	ctx := context.Background()
	for _, d := range []*Resource{vault(), number()} {
		if diags := d.ResourceSchema(ctx).ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: %v", d.TypeName, diags)
		}
	}
	if diags := setBlocklist().Schema().ValidateImplementation(ctx); diags.HasError() {
		t.Errorf("action: %v", diags)
	}
	if diags := boxName().Schema().ValidateImplementation(ctx); diags.HasError() {
		t.Errorf("ephemeral: %v", diags)
	}
}
