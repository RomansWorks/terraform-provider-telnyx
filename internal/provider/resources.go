// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/provider/generated"
)

// hooks holds the hand-written behaviour of resources whose API the spec
// cannot fully describe, by type name. See hooks_*.go.
var hooks = map[string]engine.Hooks{
	"telnyx_ai_avatar":                            aiAvatarHooks,
	"telnyx_ai_tool":                              aiToolHooks,
	"telnyx_display_identity_record":              displayIdentityRecordHooks,
	"telnyx_display_identity_record_phone_number": displayIdentityRecordPhoneNumberHooks,
	"telnyx_enterprise_reputation_number":         enterpriseReputationNumberHooks,
	"telnyx_number_order":                         numberOrderHooks,
	"telnyx_porting_order":                        portingOrderHooks,
	"telnyx_porting_order_additional_document":    portingOrderAdditionalDocumentHooks,
	"telnyx_sim_card_order":                       simCardOrderHooks,
	"telnyx_spend_limit":                          spendLimitHooks,
	"telnyx_texml_queue":                          texmlQueueHooks,
}

// descriptors returns every generated resource with its hooks attached.
func descriptors() []*engine.Resource {
	all := generated.All()
	for _, d := range all {
		if h, ok := hooks[d.TypeName]; ok {
			d.Hooks = h
		}
	}
	return all
}

func pathRoot(name string) path.Path { return path.Root(name) }
