# After create, upload the LOA and a bill and verify ownership with the hosted
# number order actions.
resource "telnyx_messaging_hosted_number_order" "main" {
  phone_numbers        = ["+13125550123"]
  messaging_profile_id = telnyx_messaging_profile.support.id
}
