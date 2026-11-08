# messaging Specification

## Purpose
Manages messaging on Telnyx: profiles and senders, short codes and mobile numbers, hosted numbers, 10DLC registration, toll-free verification, notifications, RCS and WhatsApp.

## Requirements

### Requirement: Messaging profiles and senders
The provider SHALL manage `telnyx_messaging_profile`, `telnyx_messaging_profile_auto_response` and `telnyx_alphanumeric_sender_id`, with list data sources for each, and data sources for a profile's phone numbers, short codes and alphanumeric sender ids, the account's URL shortening domains, and the messaging settings of mobile numbers. Profile secrets SHALL be sensitive. Regenerating a profile's secret SHALL be an action.

#### Scenario: Profile secret
- **WHEN** a messaging profile is planned
- **THEN** `v1_secret` is hidden in the plan output

### Requirement: Short codes and mobile numbers
The provider SHALL manage `telnyx_short_code_settings` and `telnyx_mobile_phone_number`, and list them. The API cannot delete these objects, so create SHALL adopt the existing object and destroy SHALL only remove it from state, with a warning.

#### Scenario: Adopt a short code
- **WHEN** `telnyx_short_code_settings` is applied for short code `sc1`
- **THEN** the provider sends `PATCH /short_codes/sc1` with the configured settings, and destroy leaves the short code in Telnyx with a warning

### Requirement: Hosted messaging numbers
The provider SHALL manage `telnyx_messaging_hosted_number_order` and `telnyx_messaging_hosted_number`. Destroying a hosted number SHALL keep it in Telnyx unless `release_on_destroy = true`. Sending and checking verification codes SHALL be actions. The order's document upload is multipart and SHALL stay out of scope.

#### Scenario: Keep a hosted number
- **WHEN** a hosted number without `release_on_destroy` is destroyed
- **THEN** no DELETE is sent and the number leaves state with a warning

### Requirement: 10DLC registration
The provider SHALL manage `telnyx_10dlc_brand`, `telnyx_10dlc_campaign`, `telnyx_10dlc_phone_number_campaign` and `telnyx_10dlc_shared_campaign`, identified by `brandId`, `campaignId` and `phoneNumber`. Brand and campaign descriptions SHALL say that registration is billed and that destroying a campaign deactivates it permanently. A shared campaign SHALL be adopted and only removed from state on destroy.

#### Scenario: Assign a number to a campaign
- **WHEN** `telnyx_10dlc_phone_number_campaign` is created with `phone_number = "+15551230000"` and a campaign id
- **THEN** it is identified by the phone number, imports by it, and destroy removes the assignment

#### Scenario: List brands
- **WHEN** `data "telnyx_10dlc_brands" "all" {}` is read
- **THEN** the brands of every page are in `brands`

### Requirement: 10DLC workflow
The provider SHALL offer brand vetting, revetting, SMS OTP, 2FA email, campaign appeal and sharing acceptance as actions, and the registry's lookups (enumerations, feedback, OTP status, external vettings, use case cost and qualification, operation status, OSR attributes, sharing) as data sources. Assigning a messaging profile's numbers to a campaign SHALL be an action, with data sources for the assignment task and its numbers.

#### Scenario: Appeal a campaign
- **WHEN** `action "telnyx_10dlc_campaign_appeal"` is invoked for campaign `c1` with an `appeal_reason`
- **THEN** the provider sends `POST /10dlc/campaign/c1/appeal` once

### Requirement: Toll-free verification
The provider SHALL manage `telnyx_messaging_tollfree_verification_request` and list it. The Campaign Verify authorization token SHALL be sensitive and the phone numbers a set.

#### Scenario: Submit a request
- **WHEN** a verification request is created
- **THEN** the provider sends `POST /messaging_tollfree/verification/requests` and stores the request id

### Requirement: Notifications
The provider SHALL manage `telnyx_notification_channel`, `telnyx_notification_profile` and `telnyx_notification_setting`, list them, and read the notification events and event conditions.

#### Scenario: Setting without update
- **WHEN** a notification setting's configuration changes
- **THEN** the plan replaces it, because the API has no update

### Requirement: RCS
The provider SHALL manage `telnyx_rcs_brand`, `telnyx_rcs_agent` and `telnyx_rcs_agent_test_device`, list them, read an agent's carrier approvals and a recipient's RCS capabilities, and offer brand submit, agent submit and launch, and test number invites as actions. RCS brands and agents have no delete: destroy SHALL remove them from state with a warning.

#### Scenario: Brand contacts
- **WHEN** an RCS brand is created with `contacts.brand` and imported
- **THEN** the contact is in state, and `contact_type` only accepts `BRAND`

### Requirement: WhatsApp numbers and templates
The provider SHALL manage `telnyx_whatsapp_phone_number` and `telnyx_whatsapp_message_template`. Registering a phone number SHALL start verification; verifying, resending the code, deleting a business account and deleting a profile photo SHALL be actions. The profile photo upload is multipart and SHALL stay out of scope. A message template SHALL keep its `waba_id` after import.

#### Scenario: Import a template
- **WHEN** a message template is imported and the API returns its account as `whatsapp_business_account.id`
- **THEN** `waba_id` is that id and the next plan is empty

### Requirement: WhatsApp settings
The provider SHALL manage the settings of WhatsApp business accounts and phone numbers (`telnyx_whatsapp_business_account_settings`, `_phone_number_calling_routing`, `_calling_settings`, `_conversational_components`, `_profile`) and the account's `telnyx_whatsapp_user_data`. Settings SHALL be adopted on create and only removed from state on destroy, with a warning.

#### Scenario: Destroy a profile
- **WHEN** `telnyx_whatsapp_phone_number_profile` is destroyed
- **THEN** no request is sent and the profile stays in WhatsApp
