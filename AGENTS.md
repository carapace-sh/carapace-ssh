# AGENTS.md

Development notes for `carapace-ssh` (`github.com/carapace-sh/carapace-ssh`, Go 1.23). Go library for parsing/indexing OpenSSH-related data for shell completion. Not an `ssh -G`-compatible evaluator.

## Commands

```sh
go test ./...                          # run all package tests
go test ./config ./knownhosts          # run specific packages
go vet ./...                           # static checks
go test ./config -run '^$' -fuzz=Fuzz  # run fuzz tests explicitly (also knownhosts/authorizedkeys/destination)
```

- No `Makefile`, CI workflows, or linter config files are present. Do not invent `golangci-lint`, format, or release commands.

## Architecture

- `config/` — tolerant ssh_config parser/indexer and main completion API.
  - `index.go`: `Index`, `IndexFile`, `IndexDefault`, `IndexPaths`, `IndexConfig`, `IndexReader`, `HostCandidates`, `DefaultSources`. Follows `Include`s; dedupes visited files by resolved visit key (`abs` + `evalSymlinks` when possible). `IndexFile` is explicit-path (`-F`-like) but still follows its `Include`s. Empty/default discovery uses user + system configs (`IndexConfig("")`/`HostCandidates("")`/`IndexDefault`). `IndexPaths` skips empty `Source.Path`s. `IndexReader` uses `Source` context (path for cycle identity when set; base-dir/system fallbacks).
  - `parse.go`: `parseDirectives` lowercases keywords, allows optional `=` and flexible whitespace, strips comments, preserves quotes/escapes, tolerates CRLF and best-effort `\` line continuations. `splitArgs` preserves quoted/escaped content; empty quoted args are possible.
  - `hosts.go`: `normalizeHostToken`, `isCompletableHost`, deterministic dedupe/sort. `isCompletableHost` rejects empty tokens, tokens starting with `!`, whitespace, and any of `* ? [ ] { } ( ) | + ^ $ !`.
  - `match.go`: `extractMatchHosts` extracts only `host`/`originalhost` values (comma-split); negated `host`/`originalhost` criteria do not contribute candidates. Uses `option.LookupMatchCriterion`: `MatchArgNone` skipped; `MatchArgValue` consumes one arg (inline `k=v`, spaced `=`, `=v`, empty quoted handled by tokenization); `MatchArgRest` (`exec`) treated as opaque/rest and stops extraction (never executed); `command` is treated as value arg and consumed (never executed). Unknown criteria are skipped conservatively for alignment where feasible.
  - `facts.go`: `HostFacts`, static directive association (no runtime Match evaluation). `KnownHostsFiles()` trims and drops empty/case-insensitive `none` from aggregated `UserKnownHostsFile`/`GlobalKnownHostsFile`; original values are preserved in facts.
  - `options.go`: `Source`, functional `Option`s (`WithUserConfigPath`, `WithSystemConfigPath`, `WithHomeDir`, `WithReadFile`, `WithGlob`, `WithAbs`, `WithEvalSymlinks`), `DefaultSources`.
  - `testdata/`: fixtures for parsing/indexing regressions, including layout fixtures under `layout/`.
- `knownhosts/` — tolerant known_hosts parsing/enrichment (`IndexFile`, `IndexPaths`, `IndexDefault`, `IndexReader`, `IndexFromConfigPaths`, `Merge`). Handles plain/comma-separated/bracketed `[host]:port` entries (incl. bracketed IPv6, extracted as raw host), `@cert-authority`, and `@revoked`. Hashed (`|...`) and `@revoked` hosts are excluded from completable candidates; explicit literal `@cert-authority` host tokens may be included when otherwise completable. Tilde expansion for path inputs uses configured home. Missing/unreadable files are non-fatal.
- `destination/` — tolerant SSH/SCP/SFTP destination helpers: `HostPort`, `Target`, SCP remote split/join, `ProxyJump` split/parse/join, URI/`Remote` parsing/formatting. `ParseRemote` tries URI parsing first and may fall back to tolerant SCP-style parsing; use `ParseURI` for URI-only acceptance.
- `key/` — completion-safe identity discovery/pairing and agent hint helpers (`DiscoverIdentityFiles`, `DiscoverDefaultIdentityFiles`, `DiscoverKeyPairs`, `PairIdentityFiles`, `AgentCandidates*`, `AgentSocketPaths`). Filesystem/env inspection only; no exec/network/socket probe/private key reads.
- `authorizedkeys/` — tolerant `authorized_keys` parsing/indexing (`ParseLine`, `IndexFile`, `IndexReader`, `IndexDefault`, `Entries`, `Comments`). Missing/unreadable files are non-fatal.
- `option/` — curated ssh_config keyword and Match criterion metadata for tolerant parsing/lookups (`LookupKeyword`, `Keywords`, `LookupMatchCriterion`, `IsKnownMatchCriterion`, `IsZeroArgMatchCriterion`, `MatchCriteria`). `LookupMatchCriterion` normalizes leading `!` and `criterion=value` forms. Metadata is intentionally not a full ssh_config spec.
- `resolve/` — **opt-in strict resolution** via `ssh -G` (`Host`/`HostContext` with `WithSSHPath`, `WithConfigPath`, `WithRunner`). Executes an external process; must not be used in default completion paths.
- `skills/openssh/` — local OpenSSH completion knowledge base (`SKILL.md` + `references/`); consult when behavior depends on ssh_config/Match/Include/defaults nuances.

## Key Behaviors & Gotchas

- **Completion-safe by default:** never execute `Match exec`/`Match command` or any external process; no network access in parsing/indexing APIs. `resolve` is the only opt-in exception.
- **Tolerant/best-effort:** unknown keywords/criteria, malformed lines, malformed/binary input, and missing/unreadable includes/files must not abort recoverable indexing in common cases.
- **Explicit context, injectable deps:** callers control paths/base dirs/roots. Use `Option`s to override home/config paths and FS helpers (`readFile`, `glob`, `abs`, `evalSymlinks`) for determinism/tests.
- **`-F` vs defaults:** explicit config paths are isolated (like `-F`) but still follow `Include`s. Empty/default discovery uses user + system configs. `DefaultSources` reflects configured paths.
- **Deterministic outputs:** host/candidate/path outputs are deduped and sorted where specified.
- **Host semantics:** `HostCandidates()` returns completable literal hosts only; `DeclaredHosts()` preserves declared tokens including patterns/negations before filtering. `Match host`/`originalhost` contribute candidates only when the criterion itself is not negated. `exec` stops Match host extraction (opaque rest-of-line, never executed); `command` consumes a value arg (never executed).
- **Static facts model:** best-effort, no runtime Match evaluation, not `ssh -G` parity. Facts may be associated with declared negated `Host` tokens in the current static model; `KnownHostsFiles()` aggregates from all fact-bearing host tokens.
- **Include placement:** `Include` directives inside `Host`/`Match` blocks are followed conservatively for discovery (completion-safe over-inclusion).
- **No dynamic evaluation by default:** no canonicalization, token expansion, or runtime Match condition evaluation.

## Testing

- Standard `testing` package; mostly table-driven tests plus scenario tests using `t.TempDir()` and local `writeFile` helpers/fixtures.
- `config/index_test.go` is the main behavior suite: host filtering, Match safety/alignment, Include graphs/contexts, `-F` isolation/defaults/`IndexPaths`/`IndexReader`/`DefaultSources`, syntax tolerance (CRLF/continuations/comments/quoting/malformed/binary), static facts/known-hosts aggregation, and non-fatal reads.
- Focused unit tests in other packages (`*_test.go`) cover parsing/formatting/discovery/metadata/strict-resolution behavior.
- Fuzz tests: `config/fuzz_test.go`, `knownhosts/fuzz_test.go`, `authorizedkeys/fuzz_test.go`, `destination/fuzz_test.go`. Preserve tolerance and non-execution guarantees when changing parsers.
- When modifying behavior, add/update tests (and `testdata/` fixtures where appropriate) for tolerance, include/base-dir/context, `-F` isolation, Match `exec`/`command` safety, and deterministic output.

## Additional Context

- `README.md` documents scope, safety guarantees, contract notes, and usage examples.
- `skills/openssh/SKILL.md` indexes detailed OpenSSH completion semantics; use it for ssh_config/Match/Include/defaults/known_hosts edge cases.
