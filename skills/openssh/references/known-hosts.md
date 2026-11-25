# known_hosts Parsing and Host Enrichment

Support for enriching completable hosts from OpenSSH `known_hosts` files. This extends host sources beyond ssh_config declarations.

## Goal

A known_hosts parser/indexer can expose completable host candidates and merge/dedup helpers for combining them with config-declared hosts without side effects. It should support configurable/default path discovery and tolerant, non-fatal handling of missing/unreadable files.

## Line forms to handle

| Form / element | Notes |
|----------------|------|
| **Plain host lists** | Comma-separated hostnames/patterns in the host field. |
| **Bracketed `[host]:port`** | Non-default port notation; extract the host portion for hostname completion (e.g. `[example.com]:2222` -> `example.com`, `[2001:db8::1]:2222` -> `2001:db8::1`). Port-aware forms are a consumer concern. |
| **Wildcard/negated markers** | Recognize as syntax (e.g. `*`, `?`, `!` patterns) for parsing; completable output should follow filtering rules (patterns/negations typically not literal completions). |
| **Markers** | `@cert-authority`, `@revoked` prefixes — parse/tolerate. **`@revoked` entries must not be suggested** for completion. `@cert-authority` entries may contain scope patterns and/or explicit literal host tokens; only literal, non-wildcard, non-negated, non-hashed tokens should be considered completable (policy must be explicit). |
| **Hashed host entries** | Cannot be reversed; **skip/ignore hashed hosts for completion** but parse without error. |
| **Comments / blank lines** | Tolerate and ignore. |

Field structure (hostnames, key type, key, optional comment) should be parsed robustly enough to identify the host field even when later fields vary.

## Markers and hashed hosts

- **`@revoked`**: revoked host keys must never be accepted/suggested. Exclude revoked entries from completable host candidates.
- **`@cert-authority`**: parse/tolerate; explicit literal host tokens in the host field may be completable, but wildcard/negated/hashed tokens are not. Be explicit about this policy in tests/consumers.
- **Hashed hostnames** (e.g. `|1|...`) are one-way: do not reverse them, exclude from completion, and continue parsing.

## Default and configured locations

APIs should support:
- listing completable hostnames/ports from **given files** (path-based indexing helpers may expand a leading `~` using the configured home directory)
- listing from **default locations**: `~/.ssh/known_hosts`, `/etc/ssh/ssh_known_hosts`
- integration with config-derived paths when available: `UserKnownHostsFile`, `GlobalKnownHostsFile`

Keep path discovery explicit/configurable; avoid hidden global state (consistent with [defaults.md](defaults.md)). OpenSSH token expansion (`%h`, `%u`, etc.) is not performed.

## Merge and dedup

Provide helpers to combine:
- config-declared hosts (from ssh_config index)
- known_hosts-derived hosts

Consider:
- deduplication across sources
- deterministic ordering
- optional source metadata (e.g. where a host came from) if useful to consumers, without complicating the simple API
- port handling: whether to emit `host`, `[host]:port`, or both forms depends on consumer needs (document). A reasonable policy: completable host extraction emits the host portion only (stripping brackets/ports); bracketed IPv6 hosts are returned as raw IPv6 literals. This matches config host candidates, where raw IPv6 literals may be completable but bracketed `[host]:port` tokens are not literal completion candidates. Consumers can use destination/target helpers when they need bracketed/port-qualified display or connection strings.
- config-declared `UserKnownHostsFile`/`GlobalKnownHostsFile` paths (aggregated via config index helpers, trimming whitespace and excluding empty/`none`) can be indexed by callers for enrichment. A helper for indexing such paths may trim whitespace, ignore empty/`none` values, and expand a leading `~` when home is configured. Config indexing returns declared paths as-is (no token expansion); OpenSSH token expansion is not performed.

## Safety and tolerance

- No process execution, no network access.
- Unreadable/missing known_hosts files should be non-fatal in best-effort enrichment scenarios.
- Malformed lines should not abort parsing the entire file where recoverable.

## Consumer use cases

- `ssh`, `scp`, `sftp` target completion enriched with previously seen hosts.
- Tools like `ssh-keygen -F|-R` that operate on known hostnames.

## Testing expectations

- Plain and comma-separated host lists.
- `[host]:port` and `[ipv6]:port` forms (including raw IPv6 extraction from bracketed entries).
- `@cert-authority` / `@revoked` marker lines (revoked excluded; CA literal-token policy explicit).
- Hashed hosts ignored for completion but parsed safely.
- Wildcard/negated host tokens handled per filtering policy, including mixed comma-separated lists.
- Merge/dedup with config hosts; deterministic output.

## Related

- [host-candidates.md](host-candidates.md) — config-sourced candidates and filtering/determinism rules.
- [defaults.md](defaults.md) — explicit/configurable source modeling.
- [static-facts.md](static-facts.md) — `UserKnownHostsFile`/`GlobalKnownHostsFile` extraction.
