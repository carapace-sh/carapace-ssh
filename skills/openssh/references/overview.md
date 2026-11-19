# Overview

Project scope, guarantees, and phased roadmap for `carapace-ssh`, a long-lived Go library for parsing and indexing OpenSSH-related data for shell completion.

> **Source of truth**: [plan.md](../../../plan.md). Initial driver: [carapace-bin#3935](https://github.com/carapace-sh/carapace-bin/pull/3935).

## Purpose

`carapace-ssh` is a **fresh implementation** (no dependency on or code derived from `kevinburke/ssh_config` / `carapace-sh/ssh_config`) to avoid legacy API/behavior baggage and keep licensing simple (**MIT**).

It centralizes ssh_config/include/known_hosts logic for `ssh`, `scp`, `sftp`, and related completers, replacing ad-hoc parsing in `carapace-bin/pkg/actions/net/ssh/host.go`.

## Motivation (#3935)

SSH host completion drops configured hosts when configs contain `Match exec`, `Match originalhost`, or other Match criteria because current parsing/evaluation cannot tolerate them without executing commands.

The fix must:
- be **completion-safe** (never execute `Match exec`/`Match command` or evaluate dynamic conditions)
- collect declared host candidates even when blocks are gated by non-evaluable/unsupported Match criteria
- be reusable across OpenSSH-related completers

## Goals

- **Tolerant, completion-safe ssh_config indexer/parser**
  - Extract literal host candidates from `Host` and `Match host|originalhost`.
  - Follow `Include` directives (globs, recursion, cycle detection, `~` expansion).
  - Tolerate unknown keywords/Match criteria and malformed lines without aborting the whole file where possible.
  - **Never execute** `Match exec`, `Match command`, or any external process.
- **Explicit config resolution**: model `-F`/custom path vs user+system defaults, including common drop-in include chains.
- **Shared foundation** for other SSH completion topics (known_hosts, per-host facts, option metadata).
- **Small, well-tested, dependency-light**, MIT-licensed, stable API for carapace-bin consumers.

## Non-goals (initially)

- Full, byte-for-byte compatible OpenSSH config evaluation/matching (`ssh -G` parity), canonicalization, token expansion (`%h`, `%r`, `%p`, etc.), or dynamic Match evaluation.
- Executing or simulating `Match exec`/`Match command`.
- Becoming a general-purpose SSH client/config management library beyond completion needs.

## Design principles

| Principle | Implication |
|-----------|-------------|
| **Completion-safe by default** | Parsing/indexing APIs must not spawn processes or require runtime connection context. |
| **Tolerant/best-effort** | Unknown criteria/keywords, partial lines, or unreadable includes must not prevent returning results from the rest of the config graph. |
| **Explicit context** | Callers control/provide config path(s), base dirs, and system/user roots; no hidden global state. Defaults are opt-in helpers. |
| **Separation of concerns** | Tokenize/parse → include graph/index → query/extract APIs. Strict evaluation (if ever added) is a separate, opt-in layer. |
| **Deterministic results** | Stable, deduplicated candidate output where appropriate; avoid order dependence on map iteration in public results. |
| **MIT, minimal deps** | Prefer stdlib + existing carapace libs only when they add clear value; keep APIs decoupled. |

## Proposed repository layout

| Path | Purpose |
|------|---------|
| `go.mod` | Module `github.com/carapace-sh/carapace-ssh` (Go version aligned with carapace ecosystem). |
| `LICENSE` | MIT. |
| `README.md` | Purpose, scope, completion-safety guarantees, basic usage. |
| `AGENTS.md` | Dev notes (tests, lint, design constraints: no exec, tolerant parsing). |
| `config/` (or `sshconfig/`) | ssh_config tokenizer/parser, include resolution, index, host candidate extraction (e.g. `token.go`, `parse.go`, `include.go`, `index.go`, `hosts.go`, `defaults.go`). |
| `knownhosts/` | known_hosts parsing/enrichment (phase 3). |
| `option/` (or `metadata/`) | option/keyword/Match-criterion metadata (phase 5). |
| `internal/...` | Non-public helpers as needed. |
| `*_test.go`, `testdata/` | Tests alongside packages and fixtures. |

Exact package split can be finalized at bootstrap; keep the public surface small.

## Phased plan

### Phase 0 — Bootstrap
- `go.mod`, MIT `LICENSE`, minimal `README.md`, `AGENTS.md`.
- CI/lint/test wiring (`go test ./...`, vet/golangci-lint as applicable).
- Stub packages and design note linking to plan/issues.
- Confirm module path and package naming; treat `plan.md` as a living plan.

### Phase 1 — MVP: tolerant ssh_config indexer & host candidates (unblock #3935)
- Read/scan config without evaluating runtime conditions.
- Recognize `Host`, `Match`, `Include` (case-insensitively, optional `=`, flexible whitespace).
- Collect host candidates from `Host` args and `Match host|originalhost` (comma-separated, `!`-negated forms preserved in source semantics).
- Tolerant Match handling: never execute `exec`/`command`; parse/skip known zero-arg and argument-consuming criteria to preserve token alignment; unknown criteria must not abort parsing.
- Filter completable candidates (exclude wildcards/negations/pattern chars as appropriate) while exposing raw/filtered as needed.
- Resolve `Include` with `~` expansion, configurable base dirs, globbing, recursion, cycle detection, quoted/`=` forms, non-fatal missing/unreadable matches.
- Config sources: explicit path (`-F`, no implicit default merge, but follow its includes) and default discovery helpers (`~/.ssh/config`, `/etc/ssh/ssh_config`).
- Safety: no process execution, no network access.

See [ssh-config-syntax.md](ssh-config-syntax.md), [match.md](match.md), [include.md](include.md), [defaults.md](defaults.md), [host-candidates.md](host-candidates.md).

### Phase 2 — carapace-bin integration & default config coverage
- Delegate `pkg/actions/net/ssh/ActionHosts` to `carapace-ssh`; keep behavior/contract, tag, filtering semantics.
- Remove ad-hoc parsing and `kevinburke/ssh_config` dependency path from carapace-bin when unused.
- Audit `-F` propagation in `ssh`/`scp`/`sftp` completers.
- Improve default/drop-in coverage only with clearly defined, tested rules; update/add regression tests.

See [integration.md](integration.md).

### Phase 3 — known_hosts parsing & host enrichment
- Parse known_hosts lines (plain lists, `[host]:port`, wildcard/negated markers, `@cert-authority`/`@revoked`), ignore hashed hosts for completion.
- APIs for given/default locations and merge/dedup helpers with config hosts.

See [known-hosts.md](known-hosts.md).

### Phase 4 — per-host static facts for richer completion
- Retain static directives (`HostName`, `User`, `Port`, `IdentityFile`, `CertificateFile`, `ProxyJump`, `ProxyCommand`, known_hosts file directives).
- Best-effort/static association semantics; explicit limitations vs `ssh -G`.

See [static-facts.md](static-facts.md).

### Phase 5 — option/keyword & Match metadata
- Canonical keywords/aliases/deprecated spellings, argument/value types/enums.
- Match criterion metadata (arity, list/value/rest-of-line forms) for tolerant parsing and option completion.

See [option-metadata.md](option-metadata.md).

### Phase 6+ — extensions (as needed)
- Destination/path helpers (`user@host`, `[host]:port`, scp/sftp/URI/jump forms).
- Key/agent material helpers (default key discovery, cert/public-key pairing, FIDO/PKCS11/agent-aware suggestions where side-effect-free).
- `authorized_keys` parsing for relevant tooling.
- Opt-in strict resolution via `ssh -G` (explicit opt-in only, never default completion path).

**Current progress:** destination helpers (including ProxyJump/URI forms with query/fragment preservation), fuzz tests (`config`, `knownhosts`, `destination`, `authorizedkeys`), completion-safe key discovery/pairing, base-name helpers, and agent socket path hints (`key`), tolerant `authorized_keys` parsing (`authorizedkeys`), and opt-in strict `ssh -G` resolution (`resolve`) are implemented. Further refinements and integration validation remain as needed.

## Definition of done (MVP)

- `carapace-ssh` can index ssh_config (explicit and default paths) and return host candidates including those gated by `Match exec`/`Match originalhost`/other non-evaluable criteria, without executing commands.
- Includes (glob/quoted/relative/nested/cycles) and `-F` isolation behave as specified and tested.
- carapace-bin uses the library for `net.ssh.Hosts`, and #3935’s regression scenarios pass.
- Project is MIT-licensed with no legacy ssh_config fork dependency.

## Risks and open questions

- **Include base-dir semantics:** vary by context/version (including file vs user/system ssh dir; token expansion unavailable). Define, document, test; prefer compatibility with common setups.
- **Match argument grammar:** criteria vary in arity/form; tolerant skipping must recover `host`/`originalhost` safely. Maintain criterion metadata and safe fallbacks.
- **Syntax edge cases:** line continuations, quoting/escaping, comments, odd constructs, encoding — best-effort handling.
- **Default drop-in discovery:** distro/build variation; defaults must be configurable and non-surprising.
- **API stability:** start minimal; evolve additively with clear versioning.
- **Performance:** configs usually small; prioritize correctness/clarity; cache only if profiling shows need.

## Related

- [testing.md](testing.md) — verification strategy, including safety/non-execution guarantees.
- [integration.md](integration.md) — carapace-bin adoption path.
