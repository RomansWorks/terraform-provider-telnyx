# The profile of the user the API key belongs to. There is one per user.
resource "telnyx_user_profile" "this" {
  business_name = "Acme Inc."
  country       = "US"
}
