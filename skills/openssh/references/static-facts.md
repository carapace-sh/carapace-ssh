# Host-Specific Directives

Directives that declare per-host facts in ssh_config, and what is knowable about them statically.

## The directives

| Directive | Value | Default (per ssh_config(5)) | Completion relevance |
|-----------|-------|------------------------------|----------------------|
| `HostName` | real hostname (nicknames allowed) | the name given on the command line | effective destination; TOKENS allowed |
| `User` | remote username | local user | suggest `user@host` prefixes |
| `Port` | port number | 22 | suggest nonstandard ports |
| `IdentityFile` | identity path (repeatable; list adds up) | `~/.ssh/id_*` defaults | host-aware `-i` / identity suggestions; `-cert.pub` pairing |
| `CertificateFile` | certificate path (repeatable) | none | certificate suggestions |
| `ProxyJump` | `[user@]host[:port]`, comma-separated multi-hop, ssh URI, or `none` | none | jump-host (`-J`) suggestions |
| `ProxyCommand` | command string (extends to end of line) | none | proxy context; **never execute** |
| `IdentityAgent` | agent socket path, `none`, `SSH_AUTH_SOCK`, `$VAR` | `SSH_AUTH_SOCK` behavior | agent-aware context |
| `UserKnownHostsFile` | one or more paths | `~/.ssh/known_hosts`, `~/.ssh/known_hosts2` | locate per-host known_hosts sources |
| `GlobalKnownHostsFile` | one or more paths | `/etc/ssh/ssh_known_hosts`, `/etc/ssh/ssh_known_hosts2` | locate system known_hosts sources |

Paths in these directives may use `~`, TOKENS, and environment variables (per the man page, per-directive); TOKENS and environment variables resolve at connection time only, so static extraction should preserve them as written.

`IdentityFile none` disables identity loading; `ProxyJump none`/`ProxyCommand none` disable proxying; `IdentityAgent none` disables agent use.

## What static extraction can and cannot know

- ssh applies the **first matching** `Host`/`Match` section's value for each directive; later values are ignored. A static reader that associates a directive with every block mentioning the name therefore over-approximates: multiple conflicting values may be reported where ssh would use only one.
- Whether a `Match` block applies is not statically decidable in general (`Match exec` runs commands; `canonical`/`final` passes may change results).
- `HostName` substitution affects `Match host` matching and the effective destination.
- Command-valued directives (`ProxyCommand`, plus `Match exec`, `KnownHostsCommand`, `LocalCommand`) must be treated as opaque strings — never executed.

## Useful derived queries

Completion tools commonly derive:

- the set of hosts that declare a given directive (e.g. hosts with `IdentityFile`),
- the declared values per host token (pattern or literal),
- aggregated file paths for known_hosts indexing (for `UserKnownHostsFile`/`GlobalKnownHostsFile`, trimming whitespace and skipping empty/`none` values is sensible; ssh treats `none` as disabling the file list).

## Related

- [match.md](match.md) — Match blocks without evaluation.
- [host-candidates.md](host-candidates.md) — host tokens.
- [known-hosts.md](known-hosts.md) — indexing known_hosts files declared via the above directives.
- [option-metadata.md](option-metadata.md) — keyword metadata.
