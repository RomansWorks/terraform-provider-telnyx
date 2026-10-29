# The tool's settings go in the attribute named by its type, as JSON.
resource "telnyx_ai_tool" "order_status" {
  type         = "webhook"
  display_name = "Order status"
  webhook = jsonencode({
    name        = "order_status"
    description = "Look up the status of an order by its number."
    url         = "https://example.com/orders"
    method      = "POST"
  })
}
