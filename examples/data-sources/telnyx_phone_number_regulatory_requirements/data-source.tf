# What Telnyx needs before these Swiss numbers can be used.
data "telnyx_phone_number_regulatory_requirements" "swiss" {
  filter_phone_number = "+41215470622,+41215470633"
}
