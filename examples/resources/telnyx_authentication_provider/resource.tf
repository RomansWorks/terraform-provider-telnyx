# Single sign-on through a SAML identity provider.
resource "telnyx_authentication_provider" "okta" {
  name       = "Okta"
  short_name = "acme-okta"
  active     = true

  settings = {
    idp_entity_id        = "http://www.okta.com/exk123"
    idp_sso_target_url   = "https://acme.okta.com/app/telnyx/exk123/sso/saml"
    idp_cert_fingerprint = "13:38:C7:BB:C9:FF:4A:70:38:3A:E3:D9:5C:CD:DB:2E:50:1E:80:A7"
  }
}
