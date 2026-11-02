## 1. Engine

- [x] 1.1 `List` descriptor, list data source, paging styles (`page[number]`, `page_number`, none) and `max_items`, with fake-API tests for paging, filters and limits
- [x] 1.2 Read from a collection (list pages or parent field; object or plain-value elements), with lifecycle tests: read, gone from the list, import
- [x] 1.3 Extra update endpoints with optional clear, with lifecycle tests: create with setting, change one setting, remove a setting
- [x] 1.4 Data-source-only descriptors (single object)

## 2. Generator

- [x] 2.1 Override keys `kind`, `list`, `list_name`, `paging`, `read_from`, `updates`, with strict validation
- [x] 2.2 Derive filter attributes, item attributes and paging from list operations; reject unknown paging
- [x] 2.3 Derive extra-update attributes; resources without an item read
- [x] 2.4 Golden tests for each new shape

## 3. Pilots and docs

- [x] 3.1 Add `list` to the 7 pilot overrides where the API has one (6; phone number voice settings has no list of its own); regenerate; add examples and docs
- [x] 3.2 Check the pilots' list entries in `catalog/coverage.yaml`: they were already `data_source` entries with the right names; the generated test now checks every list and extra update against coverage
- [x] 3.3 Changelog entry (`changie new`)
- [x] 3.4 `make check` passes
