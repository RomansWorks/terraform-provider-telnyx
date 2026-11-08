# Destroy removes the settings from Terraform state only.
resource "telnyx_ai_memory_namespace_settings" "support" {
  namespace = telnyx_ai_memory_namespace.example.name
  summary = {
    instructions = "Remember the customer's products and open tickets."
  }
}
