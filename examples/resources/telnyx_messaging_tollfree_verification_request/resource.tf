resource "telnyx_messaging_tollfree_verification_request" "main" {
  business_name               = "Acme Inc."
  corporate_website           = "https://acme.example.com"
  business_addr1              = "600 Congress Avenue"
  business_city               = "Austin"
  business_state              = "Texas"
  business_zip                = "78701"
  business_contact_first_name = "Jane"
  business_contact_last_name  = "Doe"
  business_contact_email      = "jane.doe@acme.example.com"
  business_contact_phone      = "+15125550100"
  entity_type                 = "PRIVATE_PROFIT"
  use_case                    = "Order Notifications"
  use_case_summary            = "Order and shipping updates for Acme customers."
  message_volume              = "10,000"
  production_message_content  = "Acme: your order 1234 has shipped. Reply STOP to opt out."
  opt_in_workflow             = "Customers opt in with a checkbox at checkout."
  additional_information      = "Messages are sent only to customers with an open order."

  opt_in_workflow_image_urls = [
    { url = "https://acme.example.com/images/opt-in.png" },
  ]

  phone_numbers = [
    { phone_number = "+18885550100" },
  ]
}
