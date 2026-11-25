# known_hosts Files

Format and semantics of OpenSSH `known_hosts` files per ssh(1)/ssh-keygen(1).

## Purpose

`known_hosts` files store host keys for host authentication, and (as a side effect) list host names previously connected to. This makes them a useful host source for completion, complementing ssh_config declarations.

## File locations and configuration

| Directive | Default |
|-----------|---------|
| `UserKnownHostsFile` | `~/.ssh/known_hosts`, `~/.ssh/known_hosts2` |
| `GlobalKnownHostsFile` | `/etc/ssh/ssh_known_hosts`, `/etc/ssh/ssh_known_hosts2` |

- `HashKnownHosts` controls whether ssh hashes host names on write (distributions often default to yes in their configs; OpenSSH's own default is no).
- `CheckHostIP yes` additionally records the host's IP address in `~/.ssh/known_hosts`.
- `ssh-keygen -F hostname` finds entries; `ssh-keygen -R hostname` removes them (including hashed entries when given the hostname).

## Line format

Each line: `[markers] hostnames keytype base64-encoded-key [comment]`

| Element | Notes |
|---------|-------|
| **markers** | Optional prefixes: `@cert-authority` and `@revoked` |
| **hostnames** | Comma-separated list of PATTERNS (may be negated with `!`, may be hashed) |
| **keytype, key, comment** | Public key material; identify the line's structure |

### Hostname field forms

- **Plain literals**: `example.com`, `192.0.2.1`, `2001:db8::1`.
- **Bracketed port form**: non-standard ports are stored as `[host]:port`, e.g. `[example.com]:2222`; IPv6 appears as `[2001:db8::1]:2222` (brackets retained, even for IPv6).
- **Patterns**: `*`, `?` wildcards and `!` negation may appear (e.g. `*.example.com`), often in `@cert-authority` scope.
- **Hashed entries**: `|1|salt|hash` — one-way hashes of hostnames. They cannot be reversed; ssh matches them by hashing the queried name. Unknown to a static reader except as opaque tokens.

### Markers

- `@cert-authority`: the line's key is trusted as a certification authority for the hosts in the hostname field (which usually contain patterns).
- `@revoked`: the key is revoked and must never be accepted. Revoked lines should not contribute host candidates.

## Completable candidates

Reasonable extraction policies:

- **Include** plain literal host tokens (and raw IPv6 literals).
- **Exclude** hashed entries (irreversible), negated tokens, wildcard patterns, and `@revoked` lines.
- **`@cert-authority`** lines: their hostname field may be entirely pattern-based; only explicit literal tokens are candidates, if at all.
- Bracketed `[host]:port` entries yield the host portion for hostname completion; reattaching a port is a display/connection-string concern (e.g. for tools that support `user@host -p port` or `[host]:port` destination forms).
- Duplicate names across files/sources should be deduplicated with deterministic ordering.

## Tolerance notes

- Files may contain comments, blank lines, malformed lines, unknown key types; tolerate and skip what cannot be understood rather than aborting the file.
- Unreadable/missing files are common; treat as non-fatal.
- `UserKnownHostsFile`/`GlobalKnownHostsFile` values may name additional files (whitespace-separated); they may reference `~` but TOKENS (e.g. `%h`) only expand at connection time.

## Related

- [ssh-config-syntax.md](ssh-config-syntax.md) — PATTERNS operators.
- [host-candidates.md](host-candidates.md) — merging with config-declared hosts and filtering policies.
- [static-facts.md](static-facts.md) — `UserKnownHostsFile`/`GlobalKnownHostsFile` extraction.
- [defaults.md](defaults.md) — default file locations.
