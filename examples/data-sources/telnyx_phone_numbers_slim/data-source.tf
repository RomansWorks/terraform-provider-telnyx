# Active numbers tagged "office", from the faster slim listing.
data "telnyx_phone_numbers_slim" "office" {
  include_tags  = true
  filter_tag    = "office"
  filter_status = "active"
}
