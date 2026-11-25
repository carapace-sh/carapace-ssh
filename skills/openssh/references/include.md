# Include Directive

Semantics of `Include` per ssh_config(5).

## Definition

`Include` includes the specified configuration file(s):

- **Multiple pathnames** may be specified in one `Include` directive.
- Each pathname may contain **glob(7) wildcards**, **TOKENS**, and **environment variables**; for user configurations, shell-like `~` references to the user's home directory are also allowed.
- Wildcards are expanded and processed in **lexical order**.
- **Relative paths are assumed to be in `~/.ssh`** if included from a user configuration file, or in `/etc/ssh` if included from the system configuration file.
- `Include` **may appear inside a `Host` or `Match` block** to perform conditional inclusion.

## What OpenSSH does not specify

- Cycle handling is not specified; a parser following includes recursively should detect them (e.g. by resolved path) to avoid infinite loops.
- Missing files make `ssh` fail; tolerant completion extractors commonly treat missing/unreadable include targets as non-fatal and continue.
- Conditional inclusion means the file's contents apply only when the enclosing block matches; a static extractor that wants over-approximation (prefer missing a candidate over guaranteed coverage) may conservatively follow such includes.

## Extraction considerations

- Resolve relative paths according to the context of the including file (user vs system), as ssh does.
- Preserve quoted paths and the `=` separator form; multiple include targets per line are valid.
- Expand `~` using the local user's home directory; leave TOKENS and environment variables unexpanded (their values depend on connection/local state).
- Process glob matches in sorted (lexical) order for deterministic results.

## Related

- [ssh-config-syntax.md](ssh-config-syntax.md) — quoting and `=` forms.
- [defaults.md](defaults.md) — user vs system contexts and entry points.
- [overview.md](overview.md) — sources and precedence.
