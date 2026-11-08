# The API cannot delete managed accounts: destroy only removes this one from
# Terraform state.
resource "telnyx_managed_account" "reseller" {
  business_name  = "Acme Reseller"
  email          = "telecom@reseller.example.com"
  rollup_billing = true
}
