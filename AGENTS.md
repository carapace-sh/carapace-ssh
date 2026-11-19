# AGENTS.md

Development notes for `carapace-ssh`.

## Constraints

- **Completion-safe by default:** never execute `Match exec`, `Match command`, or any external process. No network access in parsing/indexing APIs.
- **Tolerant/best-effort:** unknown criteria/keywords, malformed lines, or unreadable/missing includes must not abort indexing of the rest of the config graph where recoverable.
- **Explicit context:** callers control config paths/base dirs/roots; avoid hidden global state. Default discovery is opt-in via helpers.
- **Deterministic output:** public host candidate results should be deduplicated and stably ordered (sorted) where appropriate.
- **No `ssh -G` parity by default:** do not implement dynamic Match evaluation, canonicalization, or runtime token expansion in the default completion path.

## Project layout

- `config/` - ssh_config parsing, include resolution, indexing, host candidate extraction.
- `*_test.go`, `testdata/` - table-driven tests and fixtures alongside packages.
- `plan.md` - living implementation plan/phases.
- `skills/openssh/` - knowledge base for OpenSSH config/completion behavior relevant to this library.

## Development

```sh
go test ./...
go vet ./...
```

Run repository linters if/when configured. Add focused tests for parsing tolerance, include behavior, `-F` isolation, and non-execution guarantees (especially `Match exec`).
