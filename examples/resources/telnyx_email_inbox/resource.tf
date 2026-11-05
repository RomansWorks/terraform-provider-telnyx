resource "telnyx_email_inbox" "support" {
  domain_id = telnyx_email_domain.mail.id
  username  = "support"
}

resource "telnyx_email_inbox_sender_filters" "support" {
  inbox_id  = telnyx_email_inbox.support.id
  blocklist = ["spam@example.net"]
}
