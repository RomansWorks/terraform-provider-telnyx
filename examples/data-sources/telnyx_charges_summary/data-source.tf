# Charges of October 2026. The period can be at most 31 days.
data "telnyx_charges_summary" "october" {
  start_date = "2026-10-01"
  end_date   = "2026-10-31"
}
