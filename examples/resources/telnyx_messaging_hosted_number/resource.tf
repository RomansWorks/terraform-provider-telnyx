# Manage a number whose hosted number order has completed. Destroy keeps the
# number unless release_on_destroy is true.
resource "telnyx_messaging_hosted_number" "main" {
  messaging_hosted_number_id = "1293384261075731499"
  messaging_profile_id       = telnyx_messaging_profile.support.id
  tags                       = ["hosted"]
}
