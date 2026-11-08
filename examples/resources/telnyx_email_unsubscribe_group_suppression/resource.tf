resource "telnyx_email_unsubscribe_group_suppression" "someone" {
  group_id = telnyx_email_unsubscribe_group.newsletter.id
  to       = "someone@example.org"
}
