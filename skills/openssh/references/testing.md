# Testing Strategy

Verification approach for completion-safe OpenSSH configuration parsing/indexing, with emphasis on safety, tolerance, determinism, and regression coverage.

## Principles

- **Safety first**: parsing/indexing must never execute external processes or access the network.
- **Table-driven tests** per component with focused fixtures.
- **Best-effort tolerance**: unknown/malformed constructs should not prevent correct results from the rest of the config graph.
- **Determinism**: public candidate outputs should be deduplicated and stably ordered where specified.
- **Explicit context**: tests should control paths/base dirs/roots to avoid environment coupling.

## Test layers

### Unit tests

- **Syntax/tokenization/parsing**: `Host`, `Match`, `Include` recognition (case-insensitive keywords, optional `=`, flexible whitespace, comments, blank lines, quoting), CRLF and backslash line continuations (best-effort).
- **Match handling**: tolerant extraction of `host`/`originalhost`, comma-separated values, and that negated `host`/`originalhost` criteria do not contribute candidates. Zero-arg criteria (`all`, `canonical`, `final`), argument-consuming criteria, inline/spaced `=` forms, empty quoted args where relevant, `exec` treated as opaque/rest (stops extraction) and never executed, `command` treated as value arg (consumed) and never executed, unknown criteria without aborting.
- **Include resolution**: `~` expansion, relative path/base-dir rules, globbing, recursion, cycle detection, quoted/`=` forms, deterministic include processing where relevant, and non-fatal missing/unreadable/non-matching targets.
- **Host candidates**: declared extraction vs completable filtering, filtering of patterns/wildcards/negations/whitespace per the documented rules, deduplication, and stable ordering.
- **known_hosts**: plain/comma lists, bracketed `[host]:port`, markers (e.g. `@cert-authority`, `@revoked`), hashed hosts excluded from completion but parseable, malformed-line tolerance, merge/dedup behavior.
- **Static facts**: directive extraction from relevant blocks, deterministic query results, and clearly defined behavior under Match-gated or ambiguous association cases (including any documented caveat that facts may be associated with declared negated `Host` tokens in the static model).
- **Metadata**: keyword normalization/lookup and Match criterion argument-kind/form handling (zero-arg / single-value / rest-of-line) as applicable.

### Fixtures and expectation cases

- Store representative configs and known_hosts samples as focused fixtures.
- Include real-world edge cases and previously problematic inputs as regression cases.
- Use explicit expected outputs (candidate lists, facts, parsed forms) to catch behavior and ordering regressions.

### Safety and non-execution tests

**Critical:** configs containing `Match exec "..."` or `Match command "..."` must not create side effects.

Recommended approach:
- use a temp marker file path or another observable side effect in the command argument
- run indexing/extraction
- assert the marker/side effect was **not** created

Also verify that parsing/indexing does not depend on runtime connection context.

### Tolerance and failure-mode tests

| Condition | Expected behavior (typical) |
|-----------|-----------------------------|
| Unknown keyword or Match criterion | Do not abort; ignore/skip best-effort and continue. |
| Malformed/unparseable line | Skip/recover where possible; continue with remaining content. |
| Missing include target or empty glob match | Non-fatal; continue indexing. |
| Unreadable include/known_hosts file | Non-fatal in best-effort flows; continue where recoverable. |
| Include cycle | Detect and stop recursing on that edge without failing the whole index. |

### Source/context tests

- **Explicit config path isolation**: indexing a specific file must not implicitly merge unrelated user/system defaults (but must follow its includes as defined).
- **Default discovery**: test user/system entry behavior with controlled fixtures/temp directories or resolver overrides.
- **Base-dir behavior**: cover relative include resolution for the contexts/rules you document.

### Integration tests (consumer-facing)

- Verify completer/tool-facing host suggestion behavior (including explicit config propagation).
- Confirm deterministic, filtered, deduplicated output in integration contexts.
- Add end-to-end regressions for configs that combine includes, `Match` criteria, and host declarations in ways that previously caused host loss or inconsistency.

### Fuzz/property tests (optional but valuable)

- Fuzz tokenizers/parsers on malformed or untrusted input for robustness (no panics/crashes, bounded behavior as appropriate).
- Property-style checks where useful (e.g. determinism/idempotence of extraction for the same input).

## Suggested coverage checklist

| Area | Cases |
|------|-------|
| **Match safety/extraction** | `Match exec`, `Match command`, `Match host`, `Match originalhost`, combined criteria (e.g. `user`, `localnetwork`, `tagged`, `localuser`, `version`), negations, comma-separated names, `Match all` / `canonical all`. |
| **Includes** | nested, globbed, quoted paths, paths with spaces, `=` forms, relative paths, cycles, missing/unreadable targets. |
| **Sources** | explicit config isolation, default discovery behavior, configurable roots/base dirs. |
| **Syntax tolerance** | comments, blank lines, mixed-case keywords, indentation/spacing, `=` variants. |
| **known_hosts** | plain/bracketed hosts, markers, hashed hosts ignored for completion, malformed-line tolerance, merge/dedup. |
| **Static facts/metadata** | directive extraction, stable queries, keyword/criterion normalization and arity handling. |

## Test design guidance

- Prefer small, explicit fixtures over large opaque ones; name cases by behavior.
- Use temporary directories/files for filesystem-dependent tests.
- Assert both positive extraction and negative guarantees (no execution, no leakage, no unexpected wildcard candidates).
- When behavior intentionally changes (for example, newly discoverable hosts), encode the new expectation and document rationale in tests/comments as appropriate.

## Related

- [overview.md](overview.md) — guarantees and scope.
- [match.md](match.md) — non-execution and tolerance requirements.
- [include.md](include.md) — resolution and cycle behavior.
- [defaults.md](defaults.md) — source selection and isolation.
- [host-candidates.md](host-candidates.md) — filtering and determinism.
- [known-hosts.md](known-hosts.md) — known_hosts parsing and merging.
- [integration.md](integration.md) — consumer-facing contract checks.
