# Order two local numbers in Chicago without picking them. Orders cannot be
# cancelled through the API.
resource "telnyx_inexplicit_number_order" "example" {
  connection_id        = telnyx_credential_connection.office.id
  messaging_profile_id = telnyx_messaging_profile.support.id

  ordering_groups = [
    {
      country_iso               = "US"
      phone_number_type         = "local"
      count_requested           = "2"
      national_destination_code = "312"
      features                  = ["sms", "voice"]
    },
  ]
}
