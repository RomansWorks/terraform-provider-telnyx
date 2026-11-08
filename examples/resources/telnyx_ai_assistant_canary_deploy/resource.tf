# Send a tenth of the conversations to a new version of the assistant, and
# every conversation from the beta team to it.
resource "telnyx_ai_assistant_canary_deploy" "front_desk" {
  assistant_id = telnyx_ai_assistant.front_desk.id
  rules = [
    {
      match = [
        { attribute = "team", operator = "in", values = ["beta"] },
      ]
      serve = {
        version_id = var.candidate_version_id
      }
    },
    {
      match = []
      serve = {
        rollout = [
          { version_id = var.current_version_id, weight = 90 },
          { version_id = var.candidate_version_id, weight = 10 },
        ]
      }
    },
  ]
}

variable "current_version_id" {
  type = string
}

variable "candidate_version_id" {
  type = string
}
