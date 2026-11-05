## 1. Engine and generator

- [x] 1.1 `read_from.match`: further attributes the collection element must match; required, force replacement, part of the import id; tests
- [x] 1.2 `query`: attributes sent as query parameters of update and delete operations that declare them; tests
- [x] 1.3 Body unions told apart by a boolean merge into one object; tests
- [x] 1.4 Operations on the collection path take all parameters as parents; tests
- [x] 1.5 A response field made required names the object; tests
- [x] 1.6 `once: true` keeps a value only create returns; tests
- [x] 1.7 Singletons whose response has an id use the fixed id; tests
- [x] 1.8 Extra-update attributes without a clear operation are optional and computed
- [x] 1.9 `Hooks.Delete`; a not-found error from a read hook removes the object from state
- [x] 1.10 Data-source query lookups can skip a parameter

## 2. Wave 6 overrides

- [x] 2.1 11 resources: `telnyx_authentication_provider`, `telnyx_auto_recharge_preferences`, `telnyx_user_profile`, `telnyx_oauth_client`, `telnyx_managed_account`, `telnyx_spend_limit`, `telnyx_enterprise`, `telnyx_display_identity_record`, `telnyx_display_identity_record_phone_number`, `telnyx_enterprise_reputation`, `telnyx_enterprise_reputation_number`
- [x] 2.2 Their 8 list data sources and 15 read-only data sources, including `telnyx_enterprise_display_identity_records`
- [x] 2.3 17 actions: the DIR workflow, branded calling, reputation LOA replacement, ownership confirmation, refresh, remediation, email verification, managed account enable and disable, reputation number removal
- [x] 2.4 Hooks: DIR call reasons, DIR and reputation phone numbers, spend limits
- [x] 2.5 Coverage: three LOA renders (PDF) move to `out_of_scope`

## 3. Tests, docs and release notes

- [x] 3.1 Every wave 6 resource passes the conformance test; the fake API answers batches of one phone number
- [x] 3.2 `make generate` docs; examples for every wave 6 resource and the DIR submit action
- [x] 3.3 Changelog entry
- [x] 3.4 `make check` passes
