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
	"telnyx_agent_gateway":                        agentGatewayHooks,
	"telnyx_agent_gateway_key_tool_permissions":   agentGatewayToolPermissionsHooks,
	"telnyx_agent_gateway_team_tool_permissions":  agentGatewayToolPermissionsHooks,
	"telnyx_ai_avatar":                            aiAvatarHooks,
	"telnyx_ai_collection":                        aiCollectionHooks,
	"telnyx_ai_gateway_token_group":               aiGatewayHooks,
	"telnyx_ai_gateway_token_key":                 aiGatewayHooks,
	"telnyx_ai_tool":                              aiToolHooks,
	"telnyx_display_identity_record":              displayIdentityRecordHooks,
	"telnyx_display_identity_record_phone_number": displayIdentityRecordPhoneNumberHooks,
	"telnyx_enterprise_reputation_number":         enterpriseReputationNumberHooks,
	"telnyx_number_order":                         numberOrderHooks,
	"telnyx_porting_order":                        portingOrderHooks,
	"telnyx_porting_order_additional_document":    portingOrderAdditionalDocumentHooks,
	"telnyx_sim_card_order":                       simCardOrderHooks,
	"telnyx_spend_limit":                          spendLimitHooks,
	"telnyx_storage_bucket_ssl_certificate":       storageBucketSSLCertificateHooks,
	"telnyx_storage_kv_key":                       storageKVKeyHooks,
	"telnyx_storage_kv_namespace":                 storageProvisionHooks,
	"telnyx_storage_sql_database":                 storageProvisionHooks,
	"telnyx_texml_queue":                          texmlQueueHooks,
	"telnyx_voice_clone":                          voiceCloneHooks,
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
