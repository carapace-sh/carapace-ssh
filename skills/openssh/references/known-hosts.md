# known_hosts Parsing and Host Enrichment

Support for enriching completable hosts from OpenSSH `known_hosts` files (Phase 3). This extends host sources beyond ssh_config declarations.

> **Source of truth**: [plan.md](../../../plan.md), Phase 3.

## Goal

The `knownhosts` package parses known_hosts content, exposes completable host candidates, and provides merge/dedup helpers for combining them with config-declared hosts without side effects. It also supports configurable/default path discovery (`DefaultPaths`, `IndexDefault`) and tolerant, non-fatal handling of missing/unreadable files.

## Line forms to handle

| Form / element | Notes |
|----------------|------|
| **Plain host lists** | Comma-separated hostnames/patterns in the host field. |
| **Bracketed `[host]:port`** | Non-default port notation; extract host and port as appropriate for completion. |
| **Wildcard/negated markers** | Recognize as syntax (e.g. `*`, `?`, `!` patterns) for parsing; completable output should follow filtering rules (patterns/negations typically not literal completions). |
| **Markers** | `@cert-authority`, `@revoked` prefixes — parse/tolerate; decide how they affect host eligibility for completion (often still useful as known hostnames, but be explicit). |
| **Hashed host entries** | Cannot be reversed; **skip/ignore hashed hosts for completion** but parse without error. |
| **Comments / blank lines** | Tolerate and ignore. |

Field structure (hostnames, key type, key, optional comment) should be parsed robustly enough to identify the host field even when later fields vary.

## Hashed hosts

Hashed hostname entries (e.g. starting with `|1|...`) are intentionally one-way. The library must:
- detect hashed entries
- not attempt to reverse them
- exclude them from completable host suggestions
- continue parsing the rest of the file without error

## Default and configured locations

APIs should support:
- listing completable hostnames/ports from **given files**
- listing from **default locations**: `~/.ssh/known_hosts`, `/etc/ssh/ssh_known_hosts`
- later integration with config-derived paths: `UserKnownHostsFile`, `GlobalKnownHostsFile` (Phase 4 static facts / future helpers)

Keep path discovery explicit/configurable; avoid hidden global state (consistent with [defaults.md](defaults.md)).

## Merge and dedup

Provide helpers to combine:
- config-declared hosts (from ssh_config index)
- known_hosts-derived hosts

Consider:
- deduplication across sources
- deterministic ordering
- optional source metadata (e.g. where a host came from) if useful to consumers, without complicating the simple API
- port handling: whether to emit `host`, `[host]:port`, or both forms depends on consumer needs (document)

## Safety and tolerance

- No process execution, no network access.
- Unreadable/missing known_hosts files should be non-fatal in best-effort enrichment scenarios.
- Malformed lines should not abort parsing the entire file where recoverable.

## Consumer use cases

- `ssh`, `scp`, `sftp` target completion enriched with previously seen hosts.
- Tools like `ssh-keygen -F|-R` that operate on known hostnames.

## Testing expectations

- Plain and comma-separated host lists.
- `[host]:port` forms.
- `@cert-authority` / `@revoked` marker lines.
- Hashed hosts ignored for completion but parsed safely.
- Wildcard/negated host tokens handled per filtering policy.
- Merge/dedup with config hosts; deterministic output.

## Related

- [host-candidates.md](host-candidates.md) — config-sourced candidates and filtering/determinism rules.
- [defaults.md](defaults.md) — explicit/configurable source modeling.
- [static-facts.md](static-facts.md) — future `UserKnownHostsFile`/`GlobalKnownHostsFile` extraction.
