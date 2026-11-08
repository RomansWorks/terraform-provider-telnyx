# Destroy leaves the profile as it is.
resource "telnyx_whatsapp_phone_number_profile" "support" {
  phone_number = telnyx_whatsapp_phone_number.support.phone_number
  display_name = "Acme Support"
  about        = "Order help, seven days a week."
  description  = "Acme sells home goods online. Message us about orders, returns and deliveries."
  email        = "support@acme.example.com"
  website      = "https://acme.example.com"
  address      = "600 Congress Avenue, Austin, TX 78701"
  profile_id   = telnyx_messaging_profile.support.id
}
