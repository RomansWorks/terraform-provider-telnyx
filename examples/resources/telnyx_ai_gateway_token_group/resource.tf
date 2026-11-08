resource "telnyx_ai_gateway_token_group" "internal_tools" {
  name            = "internal-tools"
  allowed_models  = ["openai/gpt-4o-mini", "openai/gpt-4o"]
  max_budget      = 250
  budget_duration = "30d"
  rpm_limit       = 120
  tpm_limit       = 200000

  # Fall back to the smaller model when the larger one fails.
  fallbacks = jsonencode({
    "openai/gpt-4o" = ["openai/gpt-4o-mini"]
  })

  retries = {
    max_attempts = 2
    backoff      = "exponential"
  }

  guardrails = {
    secrets = {
      prompt   = "block"
      response = "flag"
    }
  }
}
