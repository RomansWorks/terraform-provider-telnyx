# A token a participant joins a video room with, minted on every run and
# never stored.
ephemeral "telnyx_room_client_token" "guest" {
  room_id        = telnyx_room.standup.id
  token_ttl_secs = 600
}
