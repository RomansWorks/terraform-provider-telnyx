# Destroying a campaign deactivates it permanently.
resource "telnyx_10dlc_campaign" "alerts" {
  brand_id    = telnyx_10dlc_brand.acme.id
  usecase     = "ACCOUNT_NOTIFICATION"
  description = "Account alerts for Acme customers."
  sample1     = "Acme: your order 1234 has shipped. Reply STOP to opt out."
}
