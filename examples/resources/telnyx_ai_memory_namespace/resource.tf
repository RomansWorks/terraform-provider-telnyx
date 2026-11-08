resource "telnyx_ai_memory_namespace" "example" {
  name = "support"
}

resource "telnyx_ai_memory_namespace_settings" "example" {
  namespace = telnyx_ai_memory_namespace.example.name
  summary = {
    instructions = "Remember the customer's products and open tickets."
  }
}
