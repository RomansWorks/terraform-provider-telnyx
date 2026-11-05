## 1. Engine and generator

- [x] 1.1 `force_destroy` override key: boolean delete query parameters, local `force_destroy` attribute, sent on destroy when true; tests
- [x] 1.2 List responses that are a `oneOf`/`anyOf` of page shapes read their items from the first variant with a `data` array; tests

## 2. Wave 4 overrides

- [x] 2.1 11 resources with descriptions and sets: email domain, domain webhook, inbox, inbox sender filters, dedicated IP, IP pool, IP pool IP, suppression, template, unsubscribe group, unsubscribe group suppression
- [x] 2.2 10 data sources: the lists of the wave 4 resources (8), domain DNS records and domain health
- [x] 2.3 6 actions: start, pause, resume and complete warm-up of a dedicated IP; verify a domain; rotate its DKIM key
- [x] 2.4 Read hooks: inbox username from its address, domain `inbound_enabled` from `inbound.enabled`; tests
- [x] 2.5 Coverage: dedicated IP `PATCH`, suppression CSV import and its status move to `out_of_scope`

## 3. Tests, docs and release notes

- [x] 3.1 Conformance test: every wave 4 resource passes; the fake API reads from a parent's array field
- [x] 3.2 `make generate` docs; examples for the main wave 4 resources
- [x] 3.3 Changelog entry (`changie new`)
- [x] 3.4 `make check` passes
