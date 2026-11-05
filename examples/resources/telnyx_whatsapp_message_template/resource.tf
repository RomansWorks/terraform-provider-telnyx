resource "telnyx_whatsapp_message_template" "shipped" {
  waba_id  = "123456789012345"
  name     = "order_shipped"
  category = "UTILITY"
  language = "en_US"
  components = [{
    type = "BODY"
    text = "Your order {{1}} has shipped."
    example = {
      body_text = jsonencode([["1234"]])
    }
  }]
}
