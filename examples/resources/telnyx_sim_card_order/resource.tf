# Orders cannot be cancelled, so destroy only removes the order from state.
resource "telnyx_sim_card_order" "example" {
  address_id = telnyx_address.example.id
  quantity   = 5
}
