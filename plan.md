# carapace-ssh implementation plan

## Overview

`carapace-ssh` is a new long-lived Go library for parsing and indexing OpenSSH-related data for shell completion. It is a **fresh implementation** (no dependency on or code derived from `kevinburke/ssh_config` / `carapace-sh/ssh_config`) to avoid legacy API/behavior baggage and to keep licensing simple (**MIT**).

Initial driver: [carapace-bin#3935](https://github.com/carapace-sh/carapace-bin/pull/3935) — SSH host completion drops configured hosts when configs contain `Match exec`, `Match originalhost`, or other Match criteria because current parsing/evaluation cannot tolerate them without executing commands. The fix must be **completion-safe** (never execute `Match exec`/`Match command` or evaluate dynamic conditions) and reusable across `ssh`, `scp`, `sftp`, and related completers.

## Motivation

- **Correctness:** collect declared host candidates from `Host` and `Match host|originalhost` even when blocks are gated by non-evaluable/unsupported Match criteria.
- **Reuse:** centralize ssh_config/include/known_hosts logic instead of ad-hoc parsing in `carapace-bin/pkg/actions/net/ssh/host.go`.
- **Ownership:** purpose-built, tolerant APIs for completion (indexing/extraction) rather than adapting a strict evaluator.
- **License:** clean MIT implementation, no fork/upstream license entanglement.

## Goals

- Provide a **tolerant, completion-safe** ssh_config indexer/parser:
  - Extract literal host candidates from `Host` and `Match host|originalhost`.
  - Follow `Include` directives (globs, recursion, cycle detection, `~` expansion).
  - Tolerate unknown keywords/Match criteria and malformed lines without aborting the whole file where possible.
  - **Never execute** `Match exec`, `Match command`, or any external process.
- Model default and explicit config resolution explicitly (`-F`/custom path vs user+system defaults), including common drop-in include chains.
- Grow into a shared foundation for other SSH completion topics (known_hosts, per-host facts, option metadata).
- Small, well-tested, dependency-light library with a stable API for carapace-bin consumers.

## Non-goals (initially)

- Full, byte-for-byte compatible OpenSSH config evaluation/matching (`ssh -G` parity), canonicalization, token expansion (`%h`, `%r`, `%p`, etc.), or dynamic Match evaluation.
- Executing or simulating `Match exec`/`Match command`.
- Becoming a general-purpose SSH client/config management library beyond completion needs (can be revisited if APIs stay clean).

## Design principles

- **Completion-safe by default:** parsing/indexing APIs must not spawn processes or require runtime connection context.
- **Tolerant/best-effort:** unknown criteria/keywords, partial lines, or unreadable includes should not prevent returning results from the rest of the config graph.
- **Explicit context:** callers control/provide config path(s), base directories, and system/user roots; no hidden global state. Defaults are opt-in helpers.
- **Separation of concerns:** tokenize/parse → include graph/index → query/extract APIs. Strict evaluation (if ever added) is a separate layer/opt-in.
- **Deterministic results:** stable, deduplicated candidate output where appropriate; avoid order dependence on map iteration in public results.
- **MIT, minimal deps:** prefer stdlib + existing carapace libs only when they add clear value (e.g. quoting helpers) and keep APIs decoupled.

## Repository layout (proposed)

- `go.mod` — module `github.com/carapace-sh/carapace-ssh` (Go version aligned with carapace ecosystem).
- `LICENSE` — MIT.
- `README.md` — purpose, scope, completion-safety guarantees, basic usage.
- `AGENTS.md` — dev notes (tests, lint, design constraints: no exec, tolerant parsing).
- `config/` (or `sshconfig/`) — ssh_config tokenizer/parser, include resolution, index, host candidate extraction.
  - e.g. `token.go`, `parse.go`, `include.go`, `index.go`, `hosts.go`, `defaults.go`
- `knownhosts/` — known_hosts parsing/enrichment (phase 3).
- `option/` (or `metadata/`) — option/keyword/Match-criterion metadata (phase 5).
- `internal/...` — helpers not part of public API as needed.
- Tests alongside packages (`*_test.go`) with fixtures under `testdata/`.

(Exact package split can be finalized at bootstrap; keep public surface small.)

## Phased plan

### Phase 0 — Bootstrap

**Deliverables**
- New repo initialized with `go.mod`, MIT `LICENSE`, minimal `README.md`, `AGENTS.md`.
- CI/lint/test wiring consistent with other carapace libs (e.g. `go test ./...`, `go vet`/golangci-lint as applicable).
- Empty or stub packages and a short design note linking to this plan/issues.

**Tasks**
- Confirm module path and package naming (`carapace-ssh`).
- Add `plan.md` maintenance guidance (treat as living plan; update as phases land).

### Phase 1 — MVP: tolerant ssh_config indexer & host candidates (unblock #3935)

**Goal:** replace carapace-bin’s ad-hoc host scanner with a library API that fixes Match-related host loss.

**Requirements (behavioral spec / initial acceptance)**
- Read and scan config text from a path (and/or reader) without evaluating runtime conditions.
- Recognize directives (case-insensitively) at least: `Host`, `Match`, `Include` (with optional `=` separator and flexible whitespace, per OpenSSH-style syntax).
- **Host candidates:** collect literal names/pattern tokens from:
  - `Host` arguments.
  - `Match` criteria `host` and `originalhost` (including comma-separated lists and `!`-negated forms as they appear in criteria arguments).
- **Match handling (tolerant, non-executing):**
  - Do not execute `exec`/`command`.
  - Parse/skip known zero-arg criteria (`all`, `canonical`, `final`) and criteria that consume arguments so token alignment is preserved well enough to still find `host`/`originalhost` in the same line/block context.
  - Unknown criteria/keywords must not abort parsing of the file or prevent host extraction elsewhere.
- **Candidate filtering for completion:** expose raw and/or filtered candidates; filtered set must exclude non-completable patterns/wildcards and negations (e.g. tokens containing glob/wildcard/regex-like chars or whitespace as appropriate) — compatible with current behavior (drop `* ? [ ] { } ( ) | + ^ $ !`/whitespace patterns) while correctly handling `!name` negations.
- **Include resolution:**
  - Expand `~` and resolve relative paths according to documented base-dir rules (configurable; defaults: relative to `~/.ssh` for user config context, `/etc/ssh` for system config context, or including-file dir where appropriate — define and test explicitly).
  - Expand globs; missing/unreadable matches are non-fatal.
  - Recurse with cycle detection (by resolved path).
  - Preserve quoted paths/arguments and `=` forms as needed.
- **Config sources:**
  - Explicit config path (e.g. `-F`): read only that file as the entry (do **not** implicitly merge system defaults), but follow its `Include`s.
  - Default discovery helper: user (`~/.ssh/config`) and system (`/etc/ssh/ssh_config`) entry points with their standard base dirs. (Phase 2 can expand drop-in defaults; MVP must not regress #3935 behavior.)
- **Safety:** no process execution, no network access.

**API sketch (initial, subject to refinement)**
- `type Source struct { Path string; BaseDir string; System bool }` (or options) to describe resolution context.
- `func IndexFile(path string, opts ...Option) (*Index, error)` / `IndexPaths(...)` / `IndexDefault(...)`.
- `func (idx *Index) HostCandidates() []string` (filtered, deduped, sorted) and/or `DeclaredHosts()` for richer info.
- Options for: custom base dir, include resolver overrides (for tests), read limits/symlink policy if needed.

**Tests (minimum)**
- Port/adapt cases from #3935: `Match exec`, `Match originalhost`, combined/other criteria (`user`, `localnetwork`, `tagged`, `localuser`, `command`, `version`), negated criteria/hosts, comma-separated names, `Match all`/`canonical all`.
- Includes: nested, glob, quoted paths, paths with spaces, `=` forms, relative paths, cycles, missing includes.
- Custom `-F` isolation (no system/user default leakage beyond what is specified).
- **Non-execution guarantee:** configs containing `Match exec "..."` must not create side effects (test via temp marker file / observable absence).
- Comments, blank lines, mixed case keywords, indentation/`=` variants.

**Deliverables**
- `config` package with parsing, include resolution, index, and host-candidate API.
- Comprehensive unit tests + fixtures.
- Documentation of supported syntax subset and known limitations.

### Phase 2 — carapace-bin integration & default config coverage

**Goal:** make carapace-bin consume `carapace-ssh` and improve real-world host discovery.

**Tasks**
- Replace `pkg/actions/net/ssh/host.go` implementation to delegate to `carapace-ssh` (keep `ActionHosts(config string)` behavior/contract, tag, filtering semantics).
- Remove ad-hoc config parsing and the `carapace-sh/ssh_config`/`kevinburke/ssh_config` dependency path from carapace-bin once unused.
- Audit `-F` behavior in `ssh`/`scp`/`sftp` completers (`completers/common/{ssh,scp,sftp}_completer`) — ensure custom config path is passed correctly and defaults remain sensible.
- Improve default entry/include coverage where safe and common:
  - Consider distro/user drop-ins (e.g. `/etc/ssh/ssh_config.d/*.conf`, user `config.d` patterns) only with clearly defined, tested rules aligned with OpenSSH behavior expectations; do not guess destructively.
- Update/add tests in carapace-bin (including #3935 regression tests) to target the library-backed behavior.

**Deliverables**
- carapace-bin depends on `carapace-ssh`; #3935 resolved via library.
- Documented default search behavior.

### Phase 3 — known_hosts parsing & host enrichment

**Goal:** enrich completable hosts beyond ssh_config declarations.

**Tasks**
- `knownhosts` package: parse `known_hosts` lines (plain host lists, `[host]:port`, wildcard/negated markers as syntax, markers like `@cert-authority`/`@revoked`).
- Skip/ignore hashed host entries for completion (cannot be reversed) but parse without error.
- APIs to list completable hostnames/ports from given files and from default locations (`~/.ssh/known_hosts`, `/etc/ssh/ssh_known_hosts`, later `UserKnownHostsFile`/`GlobalKnownHostsFile` if exposed).
- Provide merge/dedup helpers for combining config-declared hosts and known_hosts (with source metadata optional).

**Deliverables**
- Usable known_hosts host source for `ssh`/`scp`/`sftp` target completion and tools like `ssh-keygen -F|-R`.

### Phase 4 — per-host static facts for richer completion

**Goal:** extract useful declared directives per host/block without dynamic evaluation.

**Tasks**
- Extend index/model to retain directive facts (at least): `HostName`, `User`, `Port`, `IdentityFile`, `CertificateFile`, `ProxyJump`, `ProxyCommand`, `UserKnownHostsFile`/`GlobalKnownHostsFile` where statically present.
- Define “best-effort/static” association semantics (e.g. candidates/facts by declared block, and/or simple matching helpers) — be explicit about limitations vs `ssh -G`.
- Consumer use cases:
  - host-aware identity/certificate suggestions for `-i`/`IdentityFile`.
  - suggest `user@` / default ports / jump hosts from declared config.
  - better `-J`/`ProxyJump` target sources.

**Deliverables**
- Query APIs for static per-host facts; carapace-bin actions adopt where valuable.

### Phase 5 — option/keyword & Match metadata

**Goal:** shared source of truth for SSH config/CLI option knowledge.

**Tasks**
- `option`/metadata package: canonical keywords, aliases, deprecated spellings, argument/value types and enums (driving/validating `pkg/actions/net/ssh/ssh.go`’s `ActionOptions`).
- Match criterion metadata (arg arity, comma-list/value form, rest-of-line cases like `exec`/`command`) to improve tolerant parsing/token skipping.
- Optional generators or data files kept in sync with OpenSSH versions as needed (manual curation acceptable initially).

**Deliverables**
- Metadata APIs consumed by option completion and config parsing.

### Phase 6+ — extensions (as needed)

- **Destination/path helpers:** parse/format `user@host`, `[host]:port`, `scp`/`sftp` path/URI forms and jump specs for consistent positional completion.
- **Key/agent material helpers:** default key discovery (`~/.ssh/id_*`), cert/public-key pairing, FIDO/PKCS11/agent-aware suggestions (where feasible without side effects).
- **`authorized_keys` parsing:** for `ssh-copy-id`/server-side tooling contexts if/when relevant.
- **Opt-in strict resolution:** optional `ssh -G`-assisted path (explicitly opt-in, may execute/evaluate per OpenSSH; never the default completion path) for callers that accept its cost/semantics.

## Testing strategy

- **Table-driven unit tests** per package with focused fixtures in `testdata/`.
- **Golden/expectation cases** derived from real-world configs and reported issues (#3935).
- **Safety tests:** assert no subprocess/side effects during parsing/indexing (esp. `Match exec`).
- **Fuzz/property tests** (later) for tokenizer/parser robustness on malformed/untrusted config input.
- **Integration in carapace-bin:** regression tests for `--macro net.ssh.Hosts(...)` and completer-level behavior (`-F`, positional host completion for `ssh`/`scp`/`sftp`).

## Integration plan (carapace-bin)

- Phase 1 lands in `carapace-ssh` with API sufficient for host candidates.
- Phase 2 switches `pkg/actions/net/ssh/ActionHosts` to `carapace-ssh` and removes `kevinburke/ssh_config` usage/replace from carapace-bin.
- Keep completion behavior stable where correct (tags, filtering, dedup/sort as appropriate) and document intentional improvements (e.g. hosts now found despite Match rules).
- Downstream completers continue to use `ssh.ActionHosts(...)`; no completer-specific rewrites required beyond benefiting from the shared fix.

## Risks and open questions

- **Include base-dir semantics:** OpenSSH behavior differs by context/version (relative to including file vs user/system ssh dir; token expansion not statically available). Must define, document, and test chosen rules; prefer compatibility with common OpenSSH setups.
- **Match argument grammar:** criteria vary in arity and form; tolerant token skipping must be good enough to recover `host`/`originalhost` without over-parsing. Maintain explicit criterion metadata (phase 5) and fall back safely.
- **Syntax edge cases:** line continuations, quoting/escaping, comments, `Match`/`Host` on same line constructs (invalid but seen?), locale/encoding oddities — handle best-effort.
- **Default drop-in discovery:** system/user include patterns vary by distro/build; make defaults configurable and avoid surprising behavior changes.
- **API stability:** start minimal; evolve with additive changes and clear versioning once carapace-bin depends on it.
- **Performance:** configs are usually small; prioritize correctness and clarity. Add caching only if profiling in completion contexts shows need.

## Definition of done (MVP)

- `carapace-ssh` can index ssh_config (explicit path and default paths) and return host candidates including those gated by `Match exec`/`Match originalhost`/other non-evaluable criteria, without executing commands.
- Includes (glob/quoted/relative/nested/cycles) and `-F` isolation behave as specified and tested.
- carapace-bin uses the library for `net.ssh.Hosts`, and #3935’s regression scenarios pass.
- Project is MIT-licensed with no legacy ssh_config fork dependency.

## Status

- **Phase 0 (Bootstrap):** complete — module, MIT license, README, AGENTS, and initial config package added.
- **Phase 1 (MVP tolerant ssh_config indexer & host candidates):** initial implementation landed — config parsing, Match host/originalhost extraction without execution, Include resolution (glob/relative/quoted/`=`/tilde/cycles/non-fatal missing), explicit/default source helpers (`IndexFile`, `IndexDefault`, `IndexPaths`, `IndexReader`, `IndexConfig`, `HostCandidates`), deterministic filtered and declared host candidates, and tests including safety/non-execution and #3935-style cases.
- **Phase 2 (carapace-bin integration):** initial integration prepared in sibling `../carapace-bin` branch `carapace-ssh-hosts` (delegates `ssh.ActionHosts` to `carapace-ssh`, removes legacy ssh_config dependency, updates sftp host completion, adds regression tests). Further validation/release integration remains in carapace-bin.
- **Phase 3 (known_hosts parsing & host enrichment):** initial implementation landed — `knownhosts` package parses plain/bracketed host entries, markers, and skips hashed hosts; provides file/path/default/reader indexing and merge helpers with tests.
- **Phase 4 (per-host static facts):** initial implementation landed — `config.Index` collects best-effort static directives (`HostName`, `User`, `Port`, `IdentityFile`, `CertificateFile`, `ProxyJump`, `ProxyCommand`, `UserKnownHostsFile`, `GlobalKnownHostsFile`) with `Facts`/`HostsWithFacts` query APIs and tests (no runtime Match evaluation, not `ssh -G` parity).
- **Phase 5 (option/keyword & Match metadata):** initial implementation landed — `option` package provides keyword and Match criterion metadata; config Match parsing uses metadata for tolerant argument handling.
- **Phase 6+:** not started (extensions, opt-in strict resolution).
