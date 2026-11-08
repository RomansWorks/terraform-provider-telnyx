resource "telnyx_email_inbox_sender_filters" "support" {
  inbox_id  = telnyx_email_inbox.support.id
  allowlist = ["@partner.example.com"]
  blocklist = ["spam@example.net"]
}
