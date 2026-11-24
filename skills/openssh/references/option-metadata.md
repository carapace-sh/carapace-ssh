# Option, Keyword, and Match Metadata

Shared metadata for OpenSSH config keywords and Match criteria. Provides a source of truth for tolerant parsing, token skipping, and option/value completion reuse.

Useful metadata includes keyword/criterion normalization and lookup helpers, argument-kind metadata for Match criteria, host-extraction markers, and list/lookup APIs for config parsing and completion.

## Goals

- Centralize SSH config/CLI option knowledge for reuse across parsing and completion.
- Improve tolerant `Match` parsing by describing criterion argument shapes.
- Drive or validate option and value completion in consumers.

## Config keyword metadata

A metadata package (e.g. `option/` or `metadata/`) may describe:

| Aspect | Examples / purpose |
|--------|--------------------|
| **Canonical keywords** | `HostName`, `User`, `Port`, `IdentityFile`, `ProxyJump`, etc. |
| **Aliases / spellings** | Accepted variants where relevant. |
| **Deprecated spellings** | Recognize for tolerance/completion hints; document status. |
| **Argument/value types** | String, path, integer/port, boolean/yes-no, enum, list, command string. |
| **Enums / allowed values** | For options with fixed choices (completion and validation helpers). |

This enables consumers to offer consistent option/value suggestions and helps the parser understand which directives carry values worth retaining (see [static-facts.md](static-facts.md)).

## Match criterion metadata

Match parsing quality depends on knowing how many/what kind of arguments each criterion consumes (see [match.md](match.md)).

Metadata to capture per criterion:
- **Argument arity**: zero-arg (`all`, `canonical`, `final`), single/multi-value, or variable.
- **Value form**: plain value, comma-separated list, pattern forms; some criteria may accept empty quoted values in specific contexts.
- **Rest-of-line / command cases**: `exec` consumes command text and must be treated as opaque and never executed. `command` is a requested command/subsystem match argument (can be empty/quoted) and must never be executed; model its argument kind explicitly for safe alignment.
- **Known names**: `host`, `originalhost`, `user`, `group`, `localuser`, `localgroup`, `localnetwork`, `tagged`, `sessiontype`, `version`, `address`, `localaddress`, `ruser`, `rdomain`, `localport`, `remoteport`, `exec`, `command`, `all`, `canonical`, `final`, etc., extensible as OpenSSH evolves.

With this metadata, tolerant skipping can preserve alignment and still recover `host`/`originalhost` reliably; unknown criteria should still fall back safely.

## Relationship to parsing and completion

```
option/Match metadata
  ├─ tolerant tokenizer/parser (skip args correctly, no exec)
  ├─ host/static-fact extraction (recognize relevant directives)
  └─ CLI/config option completion (ActionOptions, value enums)
```

Metadata is a support layer, not a runtime evaluator.

## Versioning and maintenance

- OpenSSH options and Match criteria evolve across versions.
- Optional generators or data files may be used to stay in sync; **manual curation is acceptable initially**.
- Document which OpenSSH behavior/version assumptions apply where relevant.
- Prefer additive updates; avoid breaking consumers when new keywords appear (tolerance first).

## Scope boundaries

- Metadata describes syntax/shape/known values; it does not implement connection-time matching semantics.
- Keep the package lightweight and dependency-minimal.
- Avoid coupling metadata format to a specific completer implementation except through clear APIs/adapters.

## Testing expectations

- Criterion arity/form coverage for common configs and tricky cases (`exec`/`command` rest-of-line, comma lists, zero-arg criteria, mixed criteria on one line).
- Keyword normalization/lookup behavior (case-insensitivity).
- Enum/value-type helpers (if provided) with representative options.
- Parser integration tests showing improved recovery of `host`/`originalhost` when metadata is available.

## Related

- [match.md](match.md) — non-executing Match handling and token alignment.
- [ssh-config-syntax.md](ssh-config-syntax.md) — directive recognition.
- [static-facts.md](static-facts.md) — directive facts worth retaining.
- [integration.md](integration.md) — option completion reuse in completers.
