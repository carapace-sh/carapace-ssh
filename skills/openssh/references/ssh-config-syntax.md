# ssh_config Syntax and Tolerant Parsing

Parsing model for OpenSSH client configuration as needed by a completion-safe indexer. The goal is robust extraction of host/include information, not strict `ssh -G` evaluation. For Match-specific behavior see [match.md](match.md); for include graph rules see [include.md](include.md).

## Scope

The parser/indexer must read and scan config text from a path (and/or reader) **without evaluating runtime conditions** and without executing external processes. Tolerate unknown keywords/criteria and malformed lines without aborting the whole file where possible.

## Directives (minimum set)

Recognize directives **case-insensitively**:

| Directive | Role in completion indexing |
|-----------|----------------------------|
| `Host` | Introduces a block; arguments are host patterns/candidates to collect. |
| `Match` | Introduces a conditional block; extract `host`/`originalhost` criteria as candidates; tolerate other criteria. |
| `Include` | Pulls in additional config files/globs; must be followed and indexed (see [include.md](include.md)). |

Other keywords should be accepted/tolerated (and may be retained later for static facts, [static-facts.md](static-facts.md)) without causing parse failure.

## Syntax forms to tolerate

OpenSSH-style config syntax is flexible. The parser should accept, at least:

- **Optional `=` separator**: `Host foo`, `Host=foo`, `Host = foo` (and equivalents for `Match`, `Include`).
- **Flexible whitespace**: leading/trailing whitespace, tabs/spaces around keywords and separators.
- **Case-insensitive keywords**: `host`, `Host`, `HOST`, etc.
- **Comments**: `#` starts a comment (typically to end of line); do not treat commented text as directives.
- **Blank lines**: ignored.
- **Quoted arguments**: preserve quoted paths/arguments as needed, especially for `Include` and any value that may contain spaces.
- **Indentation variants**: directives may be indented within blocks in user configs.

## Tokenization considerations

A tolerant tokenizer should aim to:
- split directives from arguments while respecting quoting and `=` forms
- keep enough structure to identify `Host`/`Match`/`Include` lines reliably
- preserve original token forms where useful for raw extraction, while normalizing for keyword matching
- allow best-effort recovery after malformed lines (continue scanning subsequent lines)

Avoid assumptions that require full OpenSSH grammar fidelity unless required for the extraction goals.

## Block structure (indexing view)

For completion indexing, `Host` and `Match` lines delimit blocks/sections. The indexer needs block awareness to:
- associate host patterns with their declaration context (relevant for future static facts)
- know that `Match` criteria appear on the `Match` line itself (not as nested directives like `Host` patterns are arguments)
- continue indexing after unsupported blocks

Do not implement runtime condition evaluation; "block applies" is irrelevant to collecting declared candidates safely.

## Candidate extraction touchpoints

- **`Host` arguments**: collect literal names/pattern tokens (filtering rules in [host-candidates.md](host-candidates.md)).
- **`Match` arguments**: scan criteria for `host` and `originalhost` and collect their value tokens/lists; skip/tolerate other criteria (see [match.md](match.md)).
- **`Include` arguments**: resolve and index included files (see [include.md](include.md)).

## Error handling / tolerance

| Situation | Expected behavior |
|-----------|-------------------|
| Unknown keyword | Do not abort file; ignore or record as opaque as needed. |
| Unknown Match criterion | Do not abort; skip criterion/args best-effort so later `host`/`originalhost` can still be found. |
| Malformed/unparseable line | Best-effort: skip/recover and continue where possible. |
| Unreadable/missing include target | Non-fatal; continue with rest of graph. |
| Mixed/odd formatting | Prefer extraction over strict rejection. |

## Edge cases to plan for

- Line continuations (backslash-newline) — handle best-effort if encountered; document behavior.
- `Match`/`Host` constructs in unusual positions (invalid but seen in the wild) — best-effort.
- `=` attached to keyword vs value in different forms.
- Comments adjacent to quoted values.
- Locale/encoding oddities — avoid crashing on unexpected bytes; prefer byte/string handling consistent with host/path usage.

## Separation from evaluation

This parser is for **indexing/extraction**. It must not:
- expand connection-time tokens (`%h`, `%r`, `%p`, etc.)
- perform canonicalization
- evaluate `Match exec`, `Match command`, or dynamic/network/user-state criteria
- claim `ssh -G` parity

Strict evaluation, if ever added, belongs in a separate opt-in layer (see [overview.md](overview.md)).

## Related

- [match.md](match.md) — criteria grammar, arity, and non-execution rules.
- [include.md](include.md) — include argument/path resolution.
- [host-candidates.md](host-candidates.md) — what becomes a completable host.
- [testing.md](testing.md) — syntax/formatting regression coverage.
