# Include Resolution

Rules and constraints for following `Include` directives when indexing ssh_config for completion. Include handling must be tolerant, cycle-safe, and side-effect-free.

## Purpose

`Include` can pull host declarations and further includes from other files. The indexer must follow includes to build a complete host-candidate view, but without executing commands or failing the entire index because one include is missing.

## Supported behaviors

- **Recognize `Include` directives** case-insensitively, with optional `=` and flexible whitespace (see [ssh-config-syntax.md](ssh-config-syntax.md)).
- **Expand `~`** in include paths.
- **Resolve relative paths** according to documented base-dir rules (configurable).
- **Expand globs** in include patterns.
- **Recurse** into included files.
- **Cycle detection** by resolved path.
- **Preserve quoted paths/arguments** and `=` forms as needed.
- **Non-fatal missing/unreadable matches**: continue indexing remaining config.

## Path resolution and base directories

Relative path resolution must be **explicitly defined and tested** because OpenSSH behavior differs by context/version.

Common contexts (configurable; define and test explicitly):
- **User config context**: relative to `~/.ssh` for certain cases (commonly relevant to default user config resolution).
- **System config context**: relative to `/etc/ssh`.
- **Including-file directory**: relative includes may be resolved relative to the file containing the directive depending on context/rules chosen.

Implementation should model context via an explicit source/options type (e.g. `Source { Path, BaseDir, System }` or equivalent options) rather than hidden globals. Callers/tests must be able to control base dirs and roots.

Token expansion (e.g. `%h`, `%u`) is generally not statically available for completion-safe indexing; avoid runtime expansion. Document and test chosen static resolution rules for compatibility with common OpenSSH setups.

## Globbing

- Expand glob patterns in include arguments.
- A pattern that matches no files is not an error that should abort indexing (treat as non-fatal).
- Process matched files in a deterministic order where practical (avoid map-iteration nondeterminism in public-visible results).

## Quoting and argument forms

- Preserve and correctly interpret quoted include paths (paths with spaces).
- Accept `Include=/path/...`, `Include /path/...`, `Include "~/path with spaces/*.conf"` and similar forms.
- Multiple include targets/arguments may appear; handle according to OpenSSH-style expectations where feasible.

## Recursion and cycle detection

- Included files may themselves contain `Include` directives; recurse.
- Detect cycles using **resolved (canonicalized for identity) paths** so the same file is not indexed repeatedly in a loop.
- When a cycle is detected, stop recursing into that edge without failing the whole index.

## Error tolerance

| Condition | Behavior |
|-----------|----------|
| Include path does not exist | Non-fatal; continue. |
| Glob matches nothing | Non-fatal; continue. |
| Permission denied / unreadable file | Non-fatal; continue (best-effort). |
| Broken/odd include line | Best-effort parse/skip; do not abort entire config graph. |
| Symlinks | Policy may be defined if needed (read limits/symlink policy options mentioned in API sketch); be explicit and test. |

## Interaction with config sources

- **Explicit config path (`-F`)**: read only that file as the entry (no implicit merge of system/user defaults) but follow its `Include`s.
- **Default discovery**: index user/system entry points and their includes under appropriate base dirs. Drop-in defaults (e.g. `/etc/ssh/ssh_config.d/*.conf`) should be added only with clearly defined, tested rules.
- **Conditional placement**: OpenSSH applies `Include` conditionally inside `Host`/`Match` blocks. A completion-safe static indexer may follow such includes conservatively for host discovery (preferring over-inclusion to missed candidates), but this tradeoff must be explicit and tested.

See [defaults.md](defaults.md).

## Testing expectations

Recommended coverage:
- nested includes
- glob includes
- quoted paths and paths with spaces
- `=` forms
- relative paths
- cycles
- missing includes

Fixtures should live under `testdata/` and be table-driven where practical. See [testing.md](testing.md).

## Related

- [ssh-config-syntax.md](ssh-config-syntax.md) — directive/token parsing.
- [defaults.md](defaults.md) — entry points and base-dir contexts.
- [overview.md](overview.md) — tolerance and explicit-context principles.
