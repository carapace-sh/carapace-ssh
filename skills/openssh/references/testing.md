# Testing Approaches

General guidance for testing tools that extract completion data from OpenSSH configuration sources.

## Principles

- **Safety first**: assert that parsing/indexing never executes `Match exec`, `ProxyCommand`, `KnownHostsCommand`, or any command-valued directive. A practical technique: use a command argument that would create an observable side effect (e.g. `touch` of a temp marker) and assert the marker was **not** created.
- **Tolerance**: unknown keywords/criteria, malformed lines, unreadable or missing includes, and missing known_hosts files should not prevent results from the recoverable remainder. Encode this as expected behavior, not incidental.
- **Determinism**: outputs (candidate lists, facts) should be deduplicated and stably ordered; assert exact output ordering.
- **Controlled environment**: drive paths, home directories, and config roots explicitly (temp directories, injectable resolvers) so tests do not depend on the host machine's real `~/.ssh` or `/etc/ssh`.

## Coverage checklist

| Area | Cases |
|------|-------|
| **Syntax** | case-insensitive keywords, `=` vs whitespace separators, quoted values (incl. empty quotes), `#` comments (incl. trailing), blank lines, leading/trailing whitespace, CRLF, backslash continuations (best-effort), binary/odd bytes |
| **Host/Match extraction** | `Host` single/multiple tokens, wildcards, negation (`!`), `Match host`/`originalhost` (comma lists, `criterion=value` forms, negated criteria), zero-arg criteria (`all`, `canonical`, `final`), combined criteria, `exec` (opaque, never executed, ambiguous alignment), `command`/`tagged` with empty quoted values |
| **Include** | multiple targets, globs (lexical order), quoted paths, paths with spaces, relative paths (user vs system context), `~` expansion, conditional inclusion inside `Host`/`Match`, cycles, missing/unreadable targets |
| **Sources** | `-F`-style isolation (no leaking of user/system defaults), default user+system discovery, drop-in configs, precedence (first-value-wins) |
| **known_hosts** | plain literals, comma lists, `[host]:port` (incl. `[ipv6]:port`), `@cert-authority`/`@revoked` markers, hashed `|1|` entries (skip, don't fail), patterns/negation, malformed lines, merge/dedup across files |
| **Static facts** | directive extraction from `Host` and `Match` blocks, repeatable directives, `none` values, ambiguous association under `Match` gating |
| **Safety** | no process execution, no network access, no dependence on runtime connection context |

## Techniques

- Table-driven tests with small, named fixtures; store representative configs and known_hosts samples (real-world edge cases as regressions).
- Property-style checks where useful (e.g. idempotence/determinism of extraction).
- Fuzz parsers on untrusted input (no panics, bounded behavior) — config files come from the network, dotfiles repos, and other users.
- Assert negative guarantees explicitly (no execution, no wildcard candidates, no leaked defaults).

## Related

- [overview.md](overview.md) — guarantees and scope.
- [match.md](match.md) — non-execution requirements.
- [integration.md](integration.md) — consumer-facing checks.
