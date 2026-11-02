# The 20 most recent pending orders.
data "telnyx_number_orders" "pending" {
  filter_status = "pending"
  max_items     = 20
}
