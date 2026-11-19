# Overview

Reference overview for parsing and indexing OpenSSH client configuration in a completion-safe way. The goal is to extract useful host and configuration data for shell completion without executing commands or depending on runtime connection state.

## Purpose

OpenSSH configuration can be distributed across default files, explicit config files, and `Include` chains. A completion-oriented parser/indexer should centralize this logic and expose deterministic, reusable results for tools that complete `ssh`, `scp`, `sftp`, and related commands.

## Core guarantees

- **Completion-safe by default**: never execute `Match exec`, `Match command`, or any external process; do not require network access or runtime connection context.
- **Tolerant/best-effort**: unknown keywords/criteria, malformed lines, or unreadable includes should not prevent returning results from the rest of the config graph where recoverable.
- **Explicit context**: callers control config paths, base directories, and system/user roots; default discovery should be opt-in and configurable.
- **Deterministic results**: public results should be stable and deduplicated where appropriate, avoiding dependence on map iteration order.
- **Separation of concerns**: tokenization/parsing, include graph/indexing, and query/extraction are distinct steps. Strict evaluation belongs in a separate opt-in layer, if implemented at all.

## Goals

- **Tolerant ssh_config indexing**
  - Recognize key directives such as `Host`, `Match`, and `Include`.
  - Extract literal host candidates from `Host` arguments and from `Match host|originalhost`.
  - Follow `Include` directives with `~` expansion, configurable base directories, globbing, recursion, and cycle detection.
  - Continue best-effort indexing when encountering unknown criteria/keywords, malformed lines, or missing/unreadable includes.
- **Explicit config resolution**
  - Support an explicit config path (isolated, like `-F`) and separate default user/system discovery.
  - Follow includes from the chosen entry point according to clearly defined rules.
- **Useful completion data sources**
  - Provide host candidates, known_hosts-derived hosts, best-effort static per-host facts, and option/Match metadata as needed by consumers.
- **Small, predictable API surface**
  - Prefer dependency-light, well-tested behavior with clearly documented limitations.

## Non-goals

- Full OpenSSH evaluation parity (including `ssh -G` behavior), canonicalization, or runtime token expansion (`%h`, `%r`, `%p`, etc.).
- Dynamic Match evaluation or simulating/executing `Match exec`/`Match command`.
- General-purpose SSH client, connection, or config management functionality beyond completion-oriented data extraction.

## Common concerns

| Concern | Notes |
|---------|------|
| **Match handling** | Preserve token alignment well enough to recover `host`/`originalhost` without executing or evaluating dynamic criteria. |
| **Include semantics** | Base-directory behavior can vary by context; define, document, and test chosen rules for explicit and default sources. |
| **Host candidate filtering** | Distinguish declared/raw tokens from completable, filtered hostnames (excluding wildcards/negations/pattern-only tokens as appropriate). |
| **Static facts** | Expose declared directives as hints, not authoritative resolved configuration. |
| **known_hosts enrichment** | Merge config-derived and known_hosts-derived hosts with deduplication and deterministic ordering. |
| **Option/Match metadata** | Use keyword/criterion metadata to improve tolerant parsing and option/value completion. |

## Related

- [ssh-config-syntax.md](ssh-config-syntax.md) — tolerant parsing and directive recognition.
- [match.md](match.md) — non-executing Match handling.
- [include.md](include.md) — include traversal and path resolution.
- [defaults.md](defaults.md) — explicit vs default config sources.
- [host-candidates.md](host-candidates.md) — extraction, filtering, and determinism.
- [known-hosts.md](known-hosts.md) — known_hosts parsing and merging.
- [static-facts.md](static-facts.md) — best-effort per-host directive facts.
- [option-metadata.md](option-metadata.md) — keyword and Match criterion metadata.
- [integration.md](integration.md) — integrating results into completers.
- [testing.md](testing.md) — safety, tolerance, and regression testing.
