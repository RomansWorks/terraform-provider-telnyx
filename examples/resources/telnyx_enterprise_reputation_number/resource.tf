resource "telnyx_enterprise_reputation_number" "main" {
  enterprise_id = telnyx_enterprise_reputation.acme.enterprise_id
  phone_number  = "+13125550100"
}
