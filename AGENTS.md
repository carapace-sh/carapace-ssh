# AGENTS.md

Development notes for `carapace-ssh` (`github.com/carapace-sh/carapace-ssh`, Go 1.23). A Go library for parsing/indexing OpenSSH-related data for shell completion (not an `ssh -G`-compatible evaluator).

## Commands

```sh
go test ./...          # run all package tests
go test ./config ./knownhosts  # run specific packages
go vet ./...           # static checks
```

- No `Makefile`, CI workflows, or linter config files are present in the repo (as observed). Do not invent `golangci-lint`/format/release commands.
- Fuzz tests exist (`*_fuzz_test.go` / `fuzz_test.go` in several packages); run explicitly if needed, e.g. `go test ./config -run '^$' -fuzz=Fuzz`.

## Architecture & package layout

- `config/` — tolerant ssh_config parser/indexer and public API for completion data.
  - `index.go`: `Index`, `IndexFile`, `IndexDefault`, `IndexPaths`, `IndexConfig`, `IndexReader`, `HostCandidates(path,...)`, `DefaultSources`. Follows `Include`s and dedupes visited files via resolved visit keys. `IndexFile` is explicit-path (`-F`-like) but follows `Include`s. Empty/default discovery uses user + system configs (`IndexConfig("")`/`HostCandidates("")`/`IndexDefault`). `IndexPaths` skips empty sources and respects `Source.BaseDir`/`Source.System`. `IndexReader` uses `Source` context (path for cycle identity when set; base-dir/system fallbacks as implemented).
  - `parse.go`: `parseDirectives` lowercases keywords, allows optional `=` and flexible whitespace, strips comments, preserves quotes/escapes, tolerates CRLF and best-effort backslash line continuations; `splitArgs` preserves quoted/escaped content (empty quoted args possible).
  - `hosts.go`: `normalizeHostToken`, `isCompletableHost` (rejects empty, tokens starting with `!`, any token containing `* ? [ ] { } ( ) | + ^ $ !` or whitespace). Deterministic dedupe/sort helpers.
  - `match.go`: `extractMatchHosts` extracts only `host`/`originalhost` (comma-split); negated `host`/`originalhost` values are not extracted. Uses `option.LookupMatchCriterion`: `MatchArgNone` skipped; `MatchArgValue` consumes one arg (inline `k=v`, spaced `=`, `=v`, empty quoted handled by tokenization); `MatchArgRest` (`exec`) treated as opaque/rest and stops (never executed); `command` treated as value arg and consumed (never executed). Unknown criteria skipped conservatively for alignment where feasible. No `exec`/`command` execution.
  - `facts.go`: `HostFacts`, static association (no runtime Match evaluation). `KnownHostsFiles()` trims and drops empty/case-insensitive `none` from aggregated `UserKnownHostsFile`/`GlobalKnownHostsFile` (original values preserved in facts).
  - `options.go`: `Source`, `Option`s (`WithUserConfigPath`, `WithSystemConfigPath`, `WithHomeDir`, `WithReadFile`, `WithGlob`, `WithAbs`, `WithEvalSymlinks`), `DefaultSources`.
  - `testdata/`: parsing/indexing fixtures.
- `knownhosts/` — tolerant known_hosts parsing/enrichment (`IndexFile`, `IndexPaths`, `IndexDefault`, `IndexReader`, merge). Bracketed `[host]:port` (incl. bracketed IPv6) yields extracted host; `@cert-authority`/`@revoked` handled; hashed (`|...`) and `@revoked` excluded from completable candidates; tilde expansion in path inputs.
- `destination/` — SSH/SCP/SFTP destination helpers (host/port, user@host[:port], scp remote split/join, ProxyJump split/parse/join, URI/Remote parsing/formatting); tolerant.
- `key/` — completion-safe identity discovery/pairing/agent helpers; filesystem/env inspection only.
- `authorizedkeys/` — tolerant `authorized_keys` parsing/indexing.
- `option/` — ssh_config keyword and Match criterion metadata (`LookupKeyword`, `LookupMatchCriterion`, `IsZeroArgMatchCriterion`, `MatchCriteria`, `MatchArgKind`), including tolerant extras (e.g. `address`, `group`, `localgroup`, `localaddress`, `localport`, `rdomain`, `ruser`, `sessiontype`, `tagged`, `version`).
- `resolve/` — opt-in strict `ssh -G` resolution (`Host(...)` with overrides/runner); executes external process; not for default completion paths.
- `skills/openssh/` — OpenSSH completion knowledge base (`SKILL.md` + references).

## Key behaviors & conventions

- **Completion-safe by default:** never execute `Match exec`/`Match command` or any external process; no network access in parsing/indexing APIs. `resolve` is the only opt-in exception.
- **Tolerant/best-effort:** unknown keywords/criteria, malformed lines, missing/unreadable includes/files do not abort recoverable indexing in common cases; malformed/binary input tolerated best-effort.
- **Explicit context, injectable deps:** callers pass paths/base dirs/roots and may inject FS helpers via `Option`s.
- **`-F` vs defaults:** explicit paths isolated but follow `Include`s; empty path uses default user+system discovery.
- **Deterministic:** deduped, sorted host-related outputs where specified.
- **Host semantics:** `HostCandidates()` returns completable literals only; `DeclaredHosts()` preserves declared tokens (incl. patterns/negations). `Match host`/`originalhost` only when not negated on the criterion. `Include` directives inside `Host`/`Match` blocks are followed conservatively for discovery (explicit tradeoff). No canonicalization/token expansion/runtime Match evaluation in default paths.

## Testing patterns

- `testing`-based, table-driven plus tempdir/fixture helpers.
- `config/index_test.go` is the main suite (host filtering, Match safety/alignment, Include graphs/contexts, `-F` isolation/defaults/`IndexPaths`/`IndexReader`/`DefaultSources`, syntax tolerance including CRLF/continuations/comments/quoting/malformed/binary, facts/known-hosts aggregation, non-fatal reads).
- Focused unit tests elsewhere; fuzz tests in `config`, `knownhosts`, `authorizedkeys`, `destination`.
- When changing behavior, update/add tests (and `testdata/` as needed) for tolerance, include/base-dir/context, `-F` isolation, and `Match exec` safety.

## Additional context

- `README.md`: scope, safety guarantees, usage.
- `skills/openssh/SKILL.md`: detailed OpenSSH completion semantics.
