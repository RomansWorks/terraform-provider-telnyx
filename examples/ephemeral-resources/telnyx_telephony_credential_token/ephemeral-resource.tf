# A JWT for a WebRTC client. It is minted on every run and never stored.
ephemeral "telnyx_telephony_credential_token" "web" {
  id = telnyx_telephony_credential.web.id
}
