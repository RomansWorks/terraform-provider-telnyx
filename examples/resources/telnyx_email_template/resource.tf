resource "telnyx_email_template" "welcome" {
  name      = "welcome"
  subject   = "Welcome, {{ first_name }}"
  html_body = "<p>Hello {{ first_name }}, thanks for signing up.</p>"
  text_body = "Hello {{ first_name }}, thanks for signing up."
}
