# Testing Strategy

Verification approach for `carapace-ssh`, with emphasis on completion-safety, tolerance, and regression coverage for [carapace-bin#3935](https://github.com/carapace-sh/carapace-bin/pull/3935).

> **Source of truth**: [plan.md](../../../plan.md), Testing strategy and phase requirements.

## Principles

- **Safety first**: parsing/indexing must never execute external processes or access the network.
- **Table-driven tests** per package with focused fixtures under `testdata/`.
- **Best-effort tolerance**: unknown/malformed constructs should not prevent correct results from the rest of the config graph.
- **Determinism**: public candidate outputs should be deduplicated and stably ordered where specified.

## Test layers

### Unit tests (per package)

- Tokenizer/parser behavior for `Host`, `Match`, `Include` (case-insensitive, `=` forms, whitespace, comments, quoting).
- Match criteria handling and host extraction (`host`, `originalhost`, comma lists, negations).
- Include resolution (path/base-dir rules, globbing, recursion, cycle detection).
- Candidate filtering, dedupe, sorting.
- Later: known_hosts parsing, static fact extraction, metadata lookups.

### Fixtures and golden/expectation cases

- Store representative configs in `testdata/`.
- Include real-world and reported-issue cases (especially #3935).
- Use explicit expected candidate lists/facts to catch regressions in behavior and ordering.

### Safety / non-execution tests

**Critical:** configs containing `Match exec "..."` or `Match command "..."` must not create side effects.

Recommended approach:
- use a temp marker file path in the command argument (or another observable side effect)
- run indexing/extraction
- assert the marker/side effect was **not** created

Also verify no reliance on runtime connection context for parsing/indexing APIs.

### Fuzz/property tests (later)

- Fuzz tokenizer/parser on malformed/untrusted config input for robustness (no panics/crashes, bounded behavior as appropriate).
- Property-style checks where useful (e.g. determinism/idempotence of candidate extraction on same input).

### Integration tests in carapace-bin (Phase 2+)

- `--macro net.ssh.Hosts(...)` regressions.
- Completer-level positional host completion for `ssh`, `scp`, `sftp`.
- `-F` propagation and isolation behavior.
- End-to-end confirmation that library-backed behavior resolves #3935 scenarios.

## Minimum coverage (MVP / Phase 1)

From the plan:

| Area | Cases |
|------|-------|
| **#3935 / Match** | `Match exec`, `Match originalhost`, combined/other criteria (`user`, `localnetwork`, `tagged`, `localuser`, `command`, `version`), negated criteria/hosts, comma-separated names, `Match all` / `canonical all`. |
| **Includes** | nested, glob, quoted paths, paths with spaces, `=` forms, relative paths, cycles, missing includes. |
| **Sources** | custom `-F` isolation (no system/user default leakage beyond specified). |
| **Safety** | no side effects from `Match exec` configs. |
| **Syntax tolerance** | comments, blank lines, mixed case keywords, indentation/`=` variants. |

## Phase-specific additions

- **Phase 3 (known_hosts):** plain/comma lists, `[host]:port`, markers `@cert-authority`/`@revoked`, hashed hosts ignored for completion but parseable, merge/dedup behavior.
- **Phase 4 (static facts):** directive extraction, deterministic association/query results, behavior under Match-gated blocks as defined.
- **Phase 5 (metadata):** criterion arity/rest-of-line handling, keyword normalization, enum/value helpers if added.

## Test design guidance

- Prefer small, explicit fixtures over large opaque ones; name cases by behavior (e.g. `match_exec_no_side_effect`, `include_cycle_stops`, `f_isolated_from_defaults`).
- Use temp dirs/files for filesystem-dependent tests; control base dirs/roots via options to avoid environment coupling.
- Assert both positive extraction (hosts found) and negative guarantees (no execution, no leakage, no wildcards in filtered output where excluded).
- When documenting intentional behavior changes (e.g. hosts now discovered despite Match rules), encode them as expectations.

## Commands

Typical local verification (once module/bootstrap exists):

```sh
go test ./...
go vet ./...
```

Run linters consistent with repo tooling (e.g. golangci-lint) when configured.

## Related

- [overview.md](overview.md) — MVP definition of done.
- [match.md](match.md) — non-execution and tolerance requirements.
- [include.md](include.md) — resolution and cycle cases.
- [defaults.md](defaults.md) — `-F` isolation and discovery.
- [integration.md](integration.md) — carapace-bin regression coverage.
