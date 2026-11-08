---
page_title: "Terraform and OpenTofu"
subcategory: ""
description: |-
  Which Terraform and OpenTofu versions the provider works with, and the features each supports.
---

# Terraform and OpenTofu

The provider is published to the Terraform Registry and the OpenTofu Registry as `romansworks/telnyx`. The same `required_providers` block works in both:

```terraform
terraform {
  required_providers {
    telnyx = {
      source = "romansworks/telnyx"
    }
  }
}
```

## What works where

| Feature | Terraform | OpenTofu |
|---|---|---|
| Resources, data sources, import | 1.11 or later | 1.11 or later |
| Ephemeral resources, such as `telnyx_telephony_credential_token` | 1.11 or later | 1.11 or later |
| Write-only attributes (the `_wo` arguments) | 1.11 or later | 1.11 or later |
| Actions, such as `telnyx_sim_card_enable` | 1.14 or later | Not available |

OpenTofu has no `action` blocks and no `action_trigger` in `lifecycle`, so it rejects a configuration that declares an action. The provider still loads in OpenTofu: it ignores the action schemas. Where an action is a one-off step, such as agreeing to terms of service or confirming a porting order, OpenTofu users can do it in the Telnyx portal or with the Telnyx API.

Every release is tested with both: CI runs the provider's tests against a fake Telnyx API with Terraform and with OpenTofu 1.11 and the latest OpenTofu, and validates every example with both CLIs.
