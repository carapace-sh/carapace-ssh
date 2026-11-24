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
  - `index.go`: `Index`, `IndexFile`, `IndexDefault`, `IndexPaths`, `IndexConfig`, `IndexReader`, `HostCandidates(path,...)`, `DefaultSources`; indexing follows `Include` directives and de-dups visited files.
  - `parse.go`: directive/token parsing (`parseDirectives`, comment/quote/escape/continuation handling, arg splitting).
  - `hosts.go`: host token normalization/filtering and deterministic sorting.
  - `match.go`: tolerant `Match` host extraction (`host`/`originalhost`), criterion-aware skipping/alignment (including `sessiontype`, empty quoted args, and `criterion=value` forms); never executes `exec`/`command` (`exec` treated as opaque/rest-arg, `command` as value arg for alignment) and does not extract hosts from negated `host`/`originalhost` criteria.
  - `facts.go`: `HostFacts` and static directive collection/association (best-effort, no runtime Match evaluation). `KnownHostsFiles()` aggregates declared `UserKnownHostsFile`/`GlobalKnownHostsFile`, trimming whitespace and excluding empty/case-insensitive `none` values (declared facts preserved as-is).
  - `options.go`: `Source`, functional `Option`s (`WithUserConfigPath`, `WithSystemConfigPath`, `WithHomeDir`, `WithReadFile`, `WithGlob`, `WithAbs`, `WithEvalSymlinks`), default source resolution.
  - `config/testdata/`: fixtures for parsing/indexing regressions (e.g. `hosts.conf`, edge-case and include-graph fixtures).
- `knownhosts/` — tolerant known_hosts parsing/enrichment (`IndexFile`, `IndexPaths`, `IndexDefault`, `IndexReader`, merge helpers). Handles plain/comma-separated/bracketed `[host]:port` (including bracketed IPv6, extracted as raw host) entries and `@cert-authority`/`@revoked`; excludes hashed (`|...`) and `@revoked` hosts from completable candidates (explicit literal `@cert-authority` host tokens may be included when otherwise completable). Expands leading `~` in path-based indexing inputs using configured home.
- `destination/` — parsing/formatting helpers for SSH/SCP/SFTP destination forms: `HostPort`, `Target` (user@host[:port]), scp remote split/join, `ProxyJump` split/parse/join, URI and `Remote` parsing/formatting; tolerant, not strict validation.
- `key/` — completion-safe identity discovery/pairing and agent hint helpers (`DiscoverIdentityFiles`, `DiscoverDefaultIdentityFiles`, `DiscoverKeyPairs`, `PairIdentityFiles`, `AgentCandidates*`, `AgentSocketPaths`); filesystem/env inspection only.
- `authorizedkeys/` — tolerant `authorized_keys` parsing (`ParseLine`, `IndexFile`, `IndexReader`, `IndexDefault`, `Entries`, `Comments`).
- `option/` — keyword and Match-criterion metadata used for tolerant parsing/lookups (`LookupKeyword`, `Keywords`, `LookupMatchCriterion`, `IsZeroArgMatchCriterion`, `MatchCriteria`).
- `resolve/` — **opt-in strict resolution** via `ssh -G` (`Host(...)` with `WithSSHPath`/`WithConfigPath`/`WithRunner`). Executes an external process; must not be used in default completion paths.
- `skills/openssh/` — local OpenSSH completion knowledge base (`SKILL.md` + references) relevant to library behavior/constraints.

## Key behaviors & conventions

- **Completion-safe by default:** never execute `Match exec`/`Match command` or any external process, and no network access in parsing/indexing APIs. `resolve` is the explicit exception and is opt-in only.
- **Tolerant/best-effort:** unknown keywords/criteria, malformed lines, and unreadable/missing includes/files must not abort indexing of recoverable parts (`IndexFile`/default/knownhosts/identity discovery treat missing/unreadable paths as non-fatal in common cases).
- **Explicit context, injectable dependencies:** callers control paths/base dirs/roots; avoid hidden global state. Use `Option`s to override home/config paths and FS helpers (`readFile`, `glob`, `abs`, `evalSymlinks`) for determinism/tests.
- **`-F` vs defaults:** explicit config path/indexing is isolated (like `-F`) but still follows its `Include`s; empty/default path discovery uses user + system configs (`IndexConfig("")` / `HostCandidates("")` / `IndexDefault`).
- **Deterministic results:** host candidates and similar outputs are deduplicated and stably sorted.
- **Host extraction semantics:** `HostCandidates()` returns completable literal hosts (filters patterns/wildcards/negations as per `isCompletableHost`); `DeclaredHosts()` preserves declared tokens including patterns/negations (pre-filter). `Match host`/`originalhost` contribute candidates unless the criterion itself is negated; static facts are associated without runtime evaluation (not `ssh -G` parity). `Include` directives encountered in `Host`/`Match` blocks are followed conservatively for discovery.
- **No dynamic evaluation by default:** no canonicalization, token expansion, or runtime Match condition evaluation in default paths.

## Testing patterns

- Standard `testing` package, mostly table-driven tests plus scenario tests using `t.TempDir()` and file fixtures/helpers (e.g. `writeFile` in package tests).
- `config/index_test.go` is the main behavior suite: host filtering, `Match exec/command` non-execution/safety, criteria tolerance/alignment (including `sessiontype`, empty/quoted args, negated criteria), comma/negation/`originalhost`, `Include` resolution/globs/relative paths/tilde/quoted names/cycles/conditional placement, `-F` isolation, user+system defaults/drop-ins, reader/path/base-dir APIs, syntax tolerance (continuations/CRLF/comments/malformed/binary), static facts/known-hosts aggregation, and non-fatal read handling.
- Other packages have focused unit tests (`*_test.go`) for parsing/formatting/discovery/metadata/strict-resolution behavior.
- Fuzz tests are present in `config`, `knownhosts`, `authorizedkeys`, and `destination` (`fuzz_test.go`); preserve/extend tolerance and non-execution guarantees when changing parsers.
- When modifying behavior, add/update tests alongside the relevant package (and `testdata/` fixtures where appropriate) covering tolerance, include/base-dir/context, `-F` isolation, and `Match exec` safety.

## Additional context

- `README.md` documents public scope, safety guarantees, and basic usage examples (`config`, `knownhosts`).
- `skills/openssh/SKILL.md` indexes detailed OpenSSH completion semantics; consult it when behavior depends on ssh_config/Match/Include/defaults nuances.
