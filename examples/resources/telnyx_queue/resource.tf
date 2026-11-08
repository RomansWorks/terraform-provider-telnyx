resource "telnyx_queue" "example" {
  queue_name = "support"
  max_size   = 50
}
