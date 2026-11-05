# Design

## Context

Wave 2 APIs come from several Telnyx teams and do not share conventions. 10DLC uses camelCase fields, `records` lists and its own paging; RCS uses maps of named entries; WhatsApp mixes `{phone_number}` and `{id}` paths and has `/v2`-prefixed paths. The aim is to absorb this in generic generator rules and keep overrides declarative.

## Goals / Non-Goals

**Goals:** every wave 2 operation implemented or classified; no wave 1 resource changes behaviour.

**Non-Goals:** multipart uploads in actions; waiting for asynchronous settings (WhatsApp calling routing) before the acceptance run shows it is needed.

## Decisions

- **Records paging is a new paging style**, not a hook: four 10DLC lists and the toll-free list use it. `totalRecords` ends the read on a full last page.
- **Echo when the response says nothing.** RCS brand `contacts` and `identifiers` are maps in the response whose entries are unions; the request names the entries. Reading them back by the request schema is the only way import can see them. Where the response map's entries have fields, those win.
- **`values` for free-form reads.** Three 10DLC lookups answer with maps keyed by carrier. One JSON attribute is honest about that; users decode it with `jsondecode`.
- **Capital runs.** The old rule made `webhook_u_r_l`. No wave 1 attribute changes, which the regenerated diff confirms, so there is no compatibility concern.
- **Version prefix in the client**, not in the descriptors, so descriptors keep the spec's keys and coverage checks still match.
- **Adopt-only objects.** Short codes, mobile numbers, RCS brands and agents, shared campaigns and WhatsApp settings have no delete: they follow the existing "Objects without delete" rule. Hosted numbers have a delete that releases the number, so they use `release_on_destroy` like phone numbers.
- **Aliases.** Where the spec has aliased paths, the classification from the triage stands (for example `GET /messaging/hosted/numbers` over `/messaging_hosted_numbers`).
- **Request-only on shape mismatch.** The spec's request and response types disagree for some fields. Reading them back fails, so the generator stops reading them. The porting order's numbers keep drift detection because its hook reshapes them (`returned`).
- **Records paging trusts the count.** The toll-free API clamps the page size, so a short page is not the last when the response gives a count.
- **Idempotency-Key.** `POST /rcs/agents` requires it. Wave 5 adds it to the client for every non-GET request; until that merges, creating RCS agents may fail. The acceptance run covers it.

## Risks / Trade-offs

- **WhatsApp calling routing** answers 202 and reads may lag; create may report an inconsistent result. Acceptance run decides whether it needs a wait hook.
- **WhatsApp phone number create** returns no body; the read straight after may 404 before verification. Acceptance run decides.
- **10DLC `alt_business_id`** is accepted only by the brand's PUT update, so the generator treats it as server-set and users cannot set it yet.
- **`telnyx_10dlc_brand_sms_otp_by_reference`** cannot filter by brand id.
- Wave 5 should merge soon after wave 2 so `telnyx_rcs_agent` can be created; no release is cut in between.
- 10DLC registration costs money and campaign deactivation is permanent; the acceptance run must use mock brands where the API allows.
