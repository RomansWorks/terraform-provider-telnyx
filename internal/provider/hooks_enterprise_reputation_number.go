// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import "github.com/RomansWorks/terraform-provider-telnyx/internal/engine"

// The API registers reputation numbers in batches. Each Terraform resource
// is one number, sent as a list of one.
var enterpriseReputationNumberHooks = engine.Hooks{
	BeforeRequest: phoneNumberAsList,
}
