# The API cannot delete RCS agents, so destroy only removes the agent from
# Terraform state.
resource "telnyx_rcs_agent" "alerts" {
  brand_id     = telnyx_rcs_brand.acme.id
  display_name = "Acme Alerts"
  use_case     = "TRANSACTIONAL"
  profile_id   = telnyx_messaging_profile.support.id

  configuration = {
    basics = jsonencode({
      description = "Order and delivery updates from Acme."
      brand_color = "#0055FF"
      logo_url    = "https://acme.example.com/rcs/logo.png"
      hero_url    = "https://acme.example.com/rcs/hero.png"
      website     = { url = "https://acme.example.com", label = "Website" }
      email       = { address = "support@acme.example.com", label = "Support" }

      privacy_policy_url       = "https://acme.example.com/privacy"
      terms_and_conditions_url = "https://acme.example.com/terms"
    })

    campaign = {
      agent_overview   = "Sends order confirmations and delivery updates."
      company_overview = "Acme sells home goods online."
      message_examples = ["Your order 1234 has shipped and arrives Friday."]
      interactions = [
        { interaction_type = "ACCOUNT_ALERTS" },
      ]
      consent_settings = {
        opt_in_methods   = ["WEBSITE"]
        call_to_action   = "Tick the box at checkout to get order updates over RCS."
        opt_out_response = "You will no longer receive Acme order updates."
        help_response    = "Acme order updates. Visit acme.example.com/help."

        call_to_action_url       = "https://acme.example.com/checkout"
        call_to_action_media_url = "https://acme.example.com/rcs/opt-in.png"
      }
    }

    testing = {
      test_url = "https://acme.example.com/rcs/test-video.mp4"
    }
  }
}
