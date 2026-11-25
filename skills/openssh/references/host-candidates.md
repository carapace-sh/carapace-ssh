# Host Tokens and Completion Candidates

Where host names appear in ssh_config, and which tokens are usable as literal completion candidates.

## Where host names are declared

1. **`Host` directive arguments**: whitespace-separated pattern lists restricting the following declarations. Each token is a PATTERNS pattern (may contain `*`, `?`, negation `!`).
2. **`Match host` / `Match originalhost` values**: single entries or comma-separated pattern lists; `host` matches the (possibly substituted) target hostname, `originalhost` the command-line hostname. Negated criteria (`!host`) and negated values (`!value`) describe exclusions, not destinations.
3. **`known_hosts` files**: previously seen hosts (see [known-hosts.md](known-hosts.md)).

## Pattern vs literal tokens

Per ssh PATTERNS semantics:

- **Literal tokens** (no `*`, `?`, no `!` prefix, no whitespace) name concrete hosts and are the natural completion candidates.
- **Pattern tokens** (`*`, `?`, `!`, character classes) describe sets of hosts; they are usually not offered as literal completions, though exposing them separately can be useful for diagnostics or fact lookup.
- Negated tokens express exclusions to wildcard matches and are not destinations.

## Selection semantics to keep in mind

- **First-match-wins**: ssh applies the first `Host`/`Match` section whose condition matches, so declaration order matters for resolution but not for candidate enumeration.
- `Match host` values are matched against the hostname *after* `HostName`/`CanonicalizeHostname` substitution, whereas `Host` arguments and `Match originalhost` refer to the command-line name. When merging candidates, this distinction affects which name will actually be used at connection time (static extraction can capture both).
- A single `*` in a `Host` line is the conventional global-defaults section.

## General extraction considerations

- Enumerate literal tokens from `Host` arguments and `Match host`/`originalhost` values (not from negated criteria), filter out tokens containing pattern characters or whitespace, deduplicate across the whole config graph, and output in a stable order (e.g. sorted).
- Bracketed `[host]:port` forms are not standard in `Host` lines; raw IPv6 literals can appear as ordinary host tokens.
- Token expansion (`%h`, `%r`, ...) is not statically available; do not synthesize expanded names.
- Whether a `Match`-gated block would actually apply is undecidable statically; extracting its declared tokens over-approximates the candidate set rather than resolving it.

## Related

- [match.md](match.md) — Match criteria semantics.
- [ssh-config-syntax.md](ssh-config-syntax.md) — PATTERNS.
- [known-hosts.md](known-hosts.md) — the other major host source.
- [include.md](include.md) and [defaults.md](defaults.md) — which files contribute.
