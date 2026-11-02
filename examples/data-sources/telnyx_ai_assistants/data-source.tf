data "telnyx_ai_assistants" "all" {}

output "assistant_names" {
  value = [for a in data.telnyx_ai_assistants.all.ai_assistants : a.name]
}
