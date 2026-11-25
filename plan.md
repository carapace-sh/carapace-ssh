# OpenSSH completion edge-case implementation plan

## Goal

Improve `carapace-ssh` so it extracts completable OpenSSH data safely, tolerantly, and deterministically across real-world `ssh_config`, `Match`, `Include`, and `known_hosts` edge cases. Default parsing/indexing must remain completion-safe: **no `Match exec`/`Match command` execution, no external processes, no network access, no `ssh -G` parity/canonicalization/token expansion**. `resolve/` remains an explicit opt-in exception.

This plan is based on the current implementation/tests (`config/`, `knownhosts/`, `option/`, `destination/`, `key/`, `authorizedkeys/`, `resolve/`), local references (`AGENTS.md`, `skills/openssh/references/*`, README), OpenSSH `ssh_config(5)` semantics, upstream `readconf.c` Match/criteria cues, and known_hosts format behavior (`sshd(8)`/`hostfile.c` summary).

## Non-goals (default paths)

- Full `ssh -G` / runtime Match evaluation, canonicalization, or connection-time token expansion (`%h`, `%r`, `%p`, `%u`, `%C`, etc.).
- Executing `Match exec`/`command`, `ProxyCommand`, `LocalCommand`, `RemoteCommand`, or probing agents/networks/hosts.
- Strict OpenSSH validation/rejection of configs; prefer best-effort recovery over aborting.

## Current coverage (baseline)

- `config`: tolerant directive parsing, host/Match host extraction, `Include` follow/glob/cycle handling, `-F` vs defaults, basic line continuation/quoting/`=` forms, static facts, non-fatal reads, fuzzing, and strong `Match exec` non-execution tests.
- `knownhosts`: plain/comma/bracketed `[host]:port` parsing, marker prefixes (`@cert-authority`, `@revoked`), hashed-host skipping, merge/default path helpers, tolerance/fuzzing.
- `option`: keyword and Match-criterion metadata used for normalization/skipping (core/common set).
- Adjacent helpers: `destination`, `key`, `authorizedkeys`, `resolve` have focused parsing/discovery behavior and tests; preserve their tolerant, side-effect-free contracts.

## Implementation priorities

1. **Safety/correctness regressions first:** never execute commands, never leak `-F` isolation, never panic/crash on malformed input.
2. **Host-discovery completeness:** avoid dropping valid declared/completable hosts due to parsing/skipping/base-dir issues.
3. **Determinism and explicit semantics:** document/test intentional tradeoffs (static facts, conditional includes, revoked/CA handling, IPv6 forms).
4. **Metadata and fuzz depth:** keep tolerant parsing aligned with real OpenSSH syntax.

## Progress

- [x] Established baseline (`go test ./...`, `go vet ./...`) and committed implementation plan (`ab7a0f9`).
- [x] `8377e6a` Match/known_hosts edge-case handling:
  - Added `sessiontype` to `option.MatchCriteria` with tests.
  - Covered `Match` host extraction/alignment when `sessiontype` is present.
  - Excluded `@revoked` known_hosts entries from completable host candidates (tolerant parsing preserved) with updated tests.
- [x] `c4d9caf` Expanded edge-case coverage:
  - Added known_hosts coverage for bracketed IPv6 (`[2001:db8::1]:2222`) and mixed revoked marker lists.
  - Extended config/known_hosts fuzz seeds (`sessiontype`, revoked/IPv6/wildcard/negated mixed cases).
- [x] `2a70621` Config parsing and Match alignment fixes:
  - Preserved empty quoted argument tokens in `config` parsing (supports `Match tagged ""` / `Match command ""`).
  - Treated `Match command` as a value-consuming criterion (not rest-of-line) for correct alignment while keeping `exec` rest-arg/non-executing.
  - Excluded `"none"` (case-insensitive) from aggregated `Index.KnownHostsFiles()` paths.
  - Added tests for empty Match criteria args and `Include` inside `Host`/`Match` blocks (followed for conservative completion discovery).
- [x] Verification: `go test ./...` and `go vet ./...` pass after each change set.
- [x] Additional edge-case coverage and smoke fuzzing:
  - Added tests for `Index.KnownHostsFiles()` excluding `"none"` (declared facts preserved) and for Match `exec="..."` inline/quoted safety with continued indexing.
  - Added known_hosts marker/mixed-token policy tests (revoked excluded; CA/wildcard/negated/hashed filtered to literal hosts; IPv6 extracted raw).
  - Added config syntax/context tolerance tests (CRLF + continuation, comments/quoting, malformed/binary input, `IndexReader` include base-dir contexts).
  - Ran short fuzz smoke runs: `config` `FuzzIndexReader`, `config` `FuzzParseDirectives`, `knownhosts` `FuzzIndexReader`, `destination` `Fuzz`, `authorizedkeys` `Fuzz` (all passed). Also ran longer config/known_hosts `FuzzIndexReader` smoke runs (8s, passed).
  - Updated `README.md` to note revoked known_hosts exclusion from completable enrichment.
- [x] Match extraction correctness fix:
  - Skipped host extraction from negated `host`/`originalhost` Match criteria while preserving argument consumption/alignment, with regression tests (including inline/mixed forms).
- [x] Documentation/reference updates:
  - Updated in-repo OpenSSH skill references (`skills/openssh/references/*`) for revoked/CA known_hosts policy, Match criteria alignment (`exec`/`command`/`sessiontype`, negated criteria, empty quoted args), known_hosts path aggregation (`none` excluded), conditional `Include` tradeoff, and syntax tolerance notes.

- [x] Regression fixtures and targeted fuzz/docs pass:
  - Added config edge-case regression fixture/test (`config/testdata/edge-cases.conf`, `edge-includes/inc.conf`) covering Match/exec-command safety, negated criteria, sessiontype/empty args, known_hosts facts/`none`, and include resolution.
  - Ran targeted 15s fuzz passes: `config` `FuzzIndexReader`/`FuzzParseDirectives`, `knownhosts` `FuzzIndexReader`, `destination` `Fuzz`, `authorizedkeys` `Fuzz` (all passed, no crashes/side effects observed).
  - Updated `AGENTS.md` to reflect refined Match/known_hosts/facts/include behavior and expanded test coverage/fixtures.

- [x] Match alignment test expansion:
  - Added tests for negated non-host Match criteria (`!user`, `!localnetwork`, `!tagged`, `!sessiontype`) preserving host extraction and for canonical/final/mixed `=`/spaced criteria alignment.
  - Added tests for negated `command`/`exec` safety and inline `exec=`/`command=` forms, canonical/final standalone forms, and unknown-criterion fallback preserving later host extraction.

- [x] Additional config context/tolerance coverage:
  - Added tests for `IndexReader` default base-dir fallback (user `~/.ssh` and system config directory contexts) when `Path`/`BaseDir` are unset.
  - Added tolerance tests for unexpanded `%` token and `$ENV`-style forms in declared values/includes (non-fatal, preserved in static facts where applicable).
  - Added Match host-list test confirming negated/pattern values are not completable but remain declared with static fact association.

- [x] Match criteria and contract/documentation refinements:
  - Added `localaddress` and `group` to `option.MatchCriteria` (value args) with lookup/list tests and config host-extraction alignment/fuzz coverage.
  - Clarified conditional `Include` handling (followed conservatively for discovery), programmatic source contexts (`IndexPaths`/`IndexReader`/`IndexConfig`), and static fact association semantics in docs and skill references; added include/context parity matrix to skill docs.
  - Documented known_hosts marker/IPv6 extraction policy and explicit literal `@cert-authority` host handling in package docs, README, AGENTS, and skill references (revoked excluded, hashed skipped, bracketed hosts extracted to hostnames/raw IPv6).
  - Kept in-flight known_hosts parser/test tolerance changes passing (host field requires key type/material; bracketed host-without-port and malformed cases covered). `go test ./...` and `go vet ./...` pass.

- [x] known_hosts path handling/enrichment follow-up:
  - Added leading `~` expansion for `knownhosts.IndexFile`/`IndexPaths` using configured home (with tests and package/skill/AGENTS docs), improving use of config-declared known_hosts paths for enrichment.
  - Added config↔knownhosts integration coverage showing `config.Index.KnownHostsFiles()` (declared paths as-is, empty/`none` excluded) can be consumed by `knownhosts.IndexPaths`/`IndexFromConfigPaths` with configured home to enrich/merge hosts (including tilde paths, missing paths tolerated). `go test ./...` and `go vet ./...` pass.
  - Added `knownhosts.IndexFromConfigPaths(...)` helper and docs for config-declared UserKnownHostsFile/GlobalKnownHostsFile enrichment.

- [x] Include/base-dir context parity follow-up:
  - Fixed `config.makeSource` to honor explicit `BaseDir` for relative source paths and to expand/normalize `BaseDir` before resolving source paths.
  - Preserved explicit `BaseDir` for top-level source indexing so `IndexPaths`/file sources use configured include context; normalized `IndexReader` `BaseDir` (including `~`) before include resolution.
  - Added tests for `IndexReader` explicit/tilde `BaseDir` (with and without `Path`), `IndexPaths` relative-source and explicit `BaseDir` precedence, and config↔knownhosts enrichment integration via `Index.KnownHostsFiles()` + `knownhosts.IndexPaths`/`Merge`. Updated OpenSSH defaults/include references. `go test ./...` and `go vet ./...` pass.

### Remaining focus
- Full Match criteria audit/unknown-criterion fallback improvements and additional parsing alignment cases. *(`localaddress`, `group`, and `localgroup` added; further audit pending)*
- Broader ssh_config syntax tolerance follow-ups (additional continuation/comment/escape edge cases, explicit `%`/env handling decisions). *(trailing/CRLF continuations and EOF continuation tolerated; comment backslash-newline no longer merges next line; unexpanded `%`/`$ENV` forms covered for tolerance/static preservation; expansion policy remains out of scope)*
- Include base-dir/context parity matrix and documentation (`-F`, defaults, `IndexReader`/`IndexPaths`, user vs system, conditional include semantics). *(core behaviors, programmatic contexts, and parity matrix documented in skill refs; explicit `BaseDir` handling and relative `IndexPaths` source cases now covered; more parity cases pending)*
- Static facts contract/tests/docs (known_hosts path semantics, Match-gated association clarity). *(association, `none`/aggregation semantics, negated Host token fact association, and Match-gated static association now locked by tests; further contract tests/docs may be useful)*
- known_hosts policy/normalization follow-ups (IPv6/bracketed form consistency across packages, helper/API ergonomics for config-declared path enrichment; CA literal-host policy now explicit but may be revisited if needed). *(tilde expansion and multi-path enrichment via existing APIs covered)*
- Longer targeted fuzz runs, regression fixtures, and README/AGENTS/skill reference updates for finalized contracts.


## Work items

### 1. Match parsing and criteria alignment (`config/match.go`, `option/match.go`, `config/index_test.go`)

**Issues/edge cases**
- `exec`/`command` must remain opaque and non-executing under all quoting/escaping/tokenization shapes (including quoted empty/complex args). Current `MatchArgRest` stop is safe; verify it cannot misread command text as later criteria or create side effects.
- Argument/empty-arg semantics: `command` and `tagged` can involve empty values/special matching in OpenSSH; `all/canonical/final` combinations and standalone forms must stay tolerant.
- Criteria coverage/alignment gaps: metadata includes core criteria but is missing/lagging newer forms such as `sessiontype` (and any other value/rest/zero-arg criteria encountered). Unknown-criterion fallback is conservative but may misalign if a true rest-arg or unusual-arity criterion is unknown.
- Negated criteria tokens (`!user`, `!host`, `!exec`) and spaced/`=` forms (`criterion = value`, `criterion=` next token) must keep host extraction/alignment safe.

**Plan**
- [~] Audit `Match` criterion handling against current `ssh_config(5)`/OpenSSH behavior and classify each criterion as zero-arg, single-value, value-list, or rest-of-line/command-like for parsing purposes. *(core kinds refined: `exec` rest-arg, `command` value arg, `sessiontype`/`localaddress`/`group`/`localgroup` added, negated host/non-host handling covered, inline/negated command-exec and unknown fallback tested; full audit pending)*
- [x] Extend `option.MatchCriteria` (and lookup helpers/tests) for missing criteria relevant to safe skipping/alignment (at minimum evaluate/add `sessiontype`; include other known criteria as needed, with explicit kinds). *(sessiontype, `localaddress`, `group`, and `localgroup` added; further audit pending)*
- [x] Keep `exec`/`command` (and any command-like criteria) non-executable and conservatively consumed; prefer safety over speculative host extraction when alignment is ambiguous. *(`exec` remains rest-arg/non-executing; `command` treated as value arg for alignment, including empty quoted values)*
- [~] Improve unknown-criterion fallback only where it can be made safer/more accurate without guessing execution or dynamic semantics. *(empty inline `criterion=` and quoted-empty unknown values now consume a following separated value token when safe; broader fallback heuristics still pending)*
- [x] Add/adjust tests for: `sessiontype`, `tagged`/`command` empty or quoted values, negated criteria, mixed criteria ordering, complex quoted `exec/command` text, `canonical`/`final` combinations, and malformed/unknown criteria not blocking later `host`/`originalhost` recovery where safe. *(sessiontype, empty `tagged`/`command`, negated `host`/`originalhost` and negated non-host criteria, canonical/final/mixed alignment covered; more malformed/complex cases pending)*

**Acceptance**
- No side effects for configs containing `Match exec`/`Match command` in any tested form.
- Declared `host`/`originalhost` values are extracted when statically present and safely parseable; unsupported/ambiguous cases fail closed without aborting the file.

### 2. `ssh_config` syntax tolerance and host token handling (`config/parse.go`, `config/hosts.go`, `config/index.go`)

**Issues/edge cases**
- Line continuations are partially handled; verify behavior with CRLF, continuations adjacent to comments/quoted strings, escaped whitespace, and malformed trailing escapes.
- Quoting/comment/escape edge cases (quoted `#`, spaces/commas in values where relevant, backslash handling) must remain tolerant and must not enable execution or incorrect include/host extraction.
- Host candidate filtering excludes patterns/negations/whitespace and metacharacters; confirm the filter is intentional and consistent with desired "literal completable host" behavior (declared/raw tokens remain available via `DeclaredHosts()`).
- Token/env expansion markers (`%...`, `$VAR`) may appear in some values/paths; static parser should treat unexpandable forms as opaque/tolerant (no execution), while preserving existing `~` behavior.

**Plan**
- [x] Add focused parser tests for continuation/comment/quote/escape combinations and malformed inputs (including binary/odd bytes and truncated lines). *(CRLF/continuation, comments/quoting, malformed/binary tolerance, and fuzz coverage added; more corner cases may remain)*
- [x] Document and, if needed, refine the completable-host filter contract (what is a literal completion vs declared pattern/negation), keeping `DeclaredHosts()`/facts semantics explicit. *(config doc, OpenSSH host-candidates ref, and filter-contract tests clarify HostCandidates vs DeclaredHosts; no behavior change needed)*
- [x] Ensure unexpanded `%` tokens and environment-variable-like forms in statically handled paths/values do not crash parsing and are handled conservatively (no dynamic expansion/evaluation). *(covered by `TestUnexpandedTokensAndEnvFormsAreTolerated`; values preserved statically, missing/unexpanded includes remain non-fatal)*

**Acceptance**
- Parser/indexer remains panic-free and best-effort on malformed/unusual syntax; deterministic host outputs for identical input/options.

### 3. `Include` resolution, context parity, and graph semantics (`config/index.go`, `config/options.go`)

**Issues/edge cases**
- Base-dir/context rules need clear parity for user vs system vs explicit `-F` vs `IndexReader` sources. `ssh_config(5)` states files without absolute paths are assumed in `~/.ssh` for user config and `/etc/ssh` for system config; relative-to-including-file resolution is also important for common setups/drop-ins.
- `IndexReader` behavior depends on `Source{Path,BaseDir,System}` and fallbacks; empty/missing context should be well-defined.
- `Include` inside `Host`/`Match` blocks: OpenSSH applies conditionally, but static indexing currently follows includes when encountered (conservative over-inclusion for discovery). This tradeoff must be explicit and tested.
- Symlink/cycle identity, missing/unreadable/glob errors, quoted/globbed/multiple include args, and deterministic match ordering are mostly covered but need parity/regression coverage for real layouts.

**Plan**
- [ ] Define/document include path-resolution rules per context (default user, default system, explicit `-F`, `IndexPaths` source, `IndexReader`) including precedence/fallback behavior for relative paths.
- [~] Add tests for: system config including `ssh_config.d/*` and nested relative includes, `-F` file with relative/glob/quoted includes, `IndexReader` with/without `Path`/`BaseDir`/`System`, and user/system base-dir expectations. *(substantial coverage exists, including reader explicit/default base-dir contexts; parity matrix/documentation pending)* *(substantial coverage exists; parity matrix/documentation pending)*
- [x] Decide and codify behavior for `Include` inside `Host`/`Match` blocks (recommended: keep conservative follow for completion discovery unless it causes material correctness issues, but make it explicit and regression-tested). *(currently followed for discovery; covered by tests and documented)*
- [ ] Preserve non-fatal handling for missing/unreadable/no-match includes and deterministic processing of glob results.
- [ ] Review `~`/path handling in include args (and interactions with quoted/`=` forms); avoid expanding tokens/env in unsafe or surprising ways.

**Acceptance**
- Explicit `-F` remains isolated from default configs but follows its include graph.
- Default discovery and include traversal are predictable, cycle-safe, and covered for common user/system/drop-in layouts.

### 4. Config sources, defaults, and static facts (`config/index.go`, `config/facts.go`)

**Issues/edge cases**
- Default entry ordering and drop-in behavior must stay configurable/testable (home/user/system path overrides, FS hooks).
- Static facts are best-effort associations (active declared host tokens), not `ssh -G`-resolved effective config; ambiguity under `Match`-gated blocks is expected but should be documented and consistently tested.
- `UserKnownHostsFile`/`GlobalKnownHostsFile` may appear multiple times, include multiple paths, `none`, relative/tilde forms, or unexpanded tokens; `KnownHostsFiles()`/consumers need a clear contract.
- `IdentityAgent` special values (`none`, `SSH_AUTH_SOCK`, paths/env-indirection forms) and multi-value directives (`IdentityFile`, `CertificateFile`) should retain useful declared data without probing.

**Plan**
- [~] Add tests clarifying default source order, `-F` isolation, and fact collection across `Host`/`Match` blocks (including gated/ambiguous cases) to lock in intended behavior. *(default order, `-F` isolation, IndexPaths/DefaultSources programmatic parity, Match-gated static association, and KnownHostsFiles aggregation/none/trim/dedup contracts covered; more ambiguous/default edge cases may remain)*
- [x] Define `KnownHostsFiles()` behavior for `none`/duplicates/empty values and whether declared values are normalized or returned as-is; update docs/tests accordingly. *(`"none"` excluded from aggregated paths, case-insensitive; declared facts preserved as-is)*
- [~] Keep fact values deterministic (stable query results/order where returned) and ensure no directive parsing triggers execution or network access. *(KnownHostsFiles/HostsWithFacts determinism covered; declared list order preserved as collected; no-exec/network guarantees covered by existing safety tests)*

**Acceptance**
- Fact APIs remain useful hints for completers while explicitly non-authoritative; no behavior implies `ssh -G` parity.

### 5. `known_hosts` parsing, filtering, and enrichment (`knownhosts/*`)

**Issues/edge cases (from `sshd(8)`/`hostfile.c` semantics)**
- Field format: optional marker, hostnames, keytype, key, comment; `#`/blank ignored; malformed/truncated/invalid-marker/bad-hash/bad-key lines must be ignored non-fatally.
- Markers: `@cert-authority` and `@revoked` prefix the host field. Revoked/CA scope entries are not ordinary connection targets; extraction policy must be explicit.
- Hashed hosts (`|1|salt|hash`) are one-way and not completable.
- Hostnames field is comma-separated pattern list: split per token; `*`/`?` patterns and `!` negations are not literal completions, but positive literal siblings on the same line may be.
- Bracketed non-standard ports: `[host]:port`, `[ipv4]:port`, `[ipv6]:port` -> host extraction must be tolerant. IPv6 literals contain `:`; current known_hosts completable filter allows `:` (unlike config), so raw IPv6 may be emitted - confirm desired output form and consistency with consumers/`destination`.
- Duplicates/conflicts and distinct port-qualified entries may appear; merge should remain deduped/sorted.

**Plan**
- [x] Document/decide policy for `@revoked` and `@cert-authority` host tokens (recommended: exclude revoked hosts from completion candidates; treat CA-scope patterns conservatively and do not invent literals) and implement/tests accordingly. *(`@revoked` excluded; explicit literal `@cert-authority` host tokens may be suggested when otherwise completable; policy now documented in code/docs/tests - may revisit if stricter behavior desired)*
- [x] Add tests for: mixed comma lists with positive + wildcard + negated tokens, `@cert-authority`/`@revoked` edge cases, bracketed IPv6 with port (`[2001:db8::1]:2222`), malformed marker/hash/field lines, and unreadable/missing files. *(key cases added; more combinations pending)*
- [~] Reconcile host candidate normalization for IPv6/bracketed forms across `knownhosts`, `config`, and `destination` helpers (define canonical completable form or clearly document raw-form behavior). *(documented current contract: raw IPv6 literals can be completable; bracketed `[host]:port` normalized to raw host by knownhosts and not treated as literal config completion; destination helpers own bracketed/port formatting)*
- [ ] Ensure `IndexPaths`/`IndexDefault`/reader behavior remains deterministic and non-fatal; support enrichment from config-declared `UserKnownHostsFile`/`GlobalKnownHostsFile` paths at API/integration level (including handling of `none` and multi-path values by callers).

**Acceptance**
- Only safe, literal, non-hashed hostname candidates are suggested by default; no panics on malformed known_hosts content.

### 6. Keyword/Match metadata and helper-package consistency (`option/`, `destination/`, `key/`, `authorizedkeys/`)

**Issues/edge cases**
- `option` metadata drives tolerant parsing/skipping and public lookups; missing or misclassified keywords/criteria can reduce recovery or desync Match parsing.
- Destination parsing must keep handling SSH/SCP/SFTP/URI forms used by completers: `user@host`, `host:port`/`[ipv6]:port`, raw IPv6 ambiguity, scp `target:path` (including Windows-drive local path handling), ProxyJump hop lists, URI query/fragment/encoded paths.
- Identity and authorized_keys helpers must remain completion-safe (filesystem/env inspection only; no probing/execution), tolerant of common forms, and consistent with config/known_hosts host representations.

**Plan**
- [~] Expand `option` keyword/criterion coverage and tests as needed for parsing-critical cases (prioritize Match criteria first); keep lists sorted and normalization behavior explicit. *(`sessiontype`, `localaddress`, `group` covered; further audit pending)*
- [ ] Add targeted tests for destination/URI/ProxyJump/IPv6 forms that intersect host completion and for helper tolerance on malformed inputs.
- [ ] Preserve non-fatal behavior for missing/unreadable discovery paths and avoid adding strict validation that rejects recoverable user data.

**Acceptance**
- Metadata and helpers support the same tolerant, deterministic, side-effect-free contract as config/known_hosts indexing.

### 7. Fuzzing, regression matrix, and documentation

**Plan**
- [~] Extend fuzz seeds/corpora for high-value combinations: *(config/knownhosts seeds expanded with negated/inline command-exec, empty/unknown/canonical-final forms, localgroup, and unknown empty-inline criteria; short, 10s, and 15s targeted fuzz runs passed on core parsers; further cases pending)*
  - `Match` with `exec/command` (complex quoting), `all/canonical/final`, unknown/new criteria, `sessiontype`, negations, empty args, continuations, and `= / spaced =` forms.
  - `Include` graphs: nested, globbed, quoted, relative, `-F` isolation, reader contexts, symlink/cycle cases, missing/unreadable targets, and conditional placement in `Host`/`Match` blocks.
  - `known_hosts`: markers, hashed, mixed literal/wildcard/negated tokens, `[ipv4]:port`, `[ipv6]:port`, malformed fields/hashes/markers.
- [~] Add regression fixtures under package `testdata/` for realistic configs/layouts (user+system+drop-in, bastion/ProxyJump/IdentityFile/known_hosts declarations) with golden candidate/facts outputs. *(config edge-case fixture + golden assertions added; broader layouts pending)*
- [x] Run and keep passing: `go test ./...`, `go vet ./...`, and targeted fuzz runs for changed parsers (e.g. `go test ./config -run '^$' -fuzz=Fuzz`, similarly for `knownhosts`, `destination`, `authorizedkeys` as relevant). *(targeted 15s fuzz runs passed for config IndexReader/ParseDirectives, knownhosts IndexReader, destination, authorizedkeys; longer campaigns pending)*
- [~] Update docs when behavior/contracts become explicit: `README.md` scope/guarantees, `AGENTS.md` notes, and `skills/openssh/references/*` (Match criteria coverage, include context rules, known_hosts host extraction policy, IPv6/`none` handling). *(README, AGENTS, and in-repo skill references updated for implemented behavior; final contract pass pending)*

**Acceptance**
- New edge cases are locked by deterministic tests/fixtures; fuzzers cover the risky parsing surfaces without finding panics or side effects.

## Suggested implementation order

1. `option` + `config/match`: Match criteria metadata/alignment and non-execution-safe parsing tests.
2. `config/parse` + syntax/continuation/quote tolerance tests.
3. `config/index` include/context parity (`-F`, defaults, reader, system/user/drop-ins, conditional include decision) with regression fixtures.
4. `config/facts` contract clarifications/tests (including known_hosts path aggregation semantics).
5. `knownhosts` extraction policy (markers/IPv6/mixed tokens) and enrichment consistency.
6. Helper-package touch-ups (`destination/key/authorizedkeys`) only where needed for consistency.
7. Fuzz seed expansion, full test/vet pass, and documentation updates.

## Definition of done

- All identified safety guarantees hold under new tests/fuzz cases (especially `Match exec/command`).
- Host candidate and known_hosts extraction behavior is explicit, deterministic, and improved for documented edge cases.
- `-F` isolation, default discovery, and include/cycle handling behave as specified.
- `go test ./...` and `go vet ./...` pass; targeted fuzz runs for modified parsers show no crashes/side effects in exercised cases.
- README/AGENTS/skill references reflect any intentional contract decisions.