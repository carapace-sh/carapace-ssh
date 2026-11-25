# Integration into Completers

Guidance for using OpenSSH configuration knowledge in shell completers for `ssh`, `scp`, `sftp`, and related tools.

## Common integration points

- **Destination/positional completion** for `ssh`/`scp`/`sftp`: host suggestions from ssh_config declarations and known_hosts entries.
- **`-F` / config selection**: flags that select an explicit configuration file should influence host discovery the same way `ssh -F` does: use only that file plus its includes, without merging system/user defaults.
- **User completion** (`user@host` prefix): `User` directives and command-line patterns can hint at usernames.
- **Identity completion** (`-i`): `IdentityFile`/`CertificateFile` declarations per host.
- **Jump host completion** (`-J` / `ProxyJump`): declared jump hosts; `ProxyJump` values are comma-separated chains of `[user@]host[:port]`.
- **Option/value completion**: keyword and criterion metadata (see [option-metadata.md](option-metadata.md)).
- **Destination forms**: ssh destinations are `[user@]hostname[:port]` (ssh), scp/sftp remote specs (`user@host:path`), and ssh URIs; port-qualified forms use `[host]:port` notation (brackets retained for IPv6).

## Behavioral expectations

- **No side effects**: parsing config or known_hosts files must never execute commands (`Match exec`, `ProxyCommand`, `KnownHostsCommand`, ...) or access the network; these values expand/run at connection time only.
- **Deterministic output**: deduplicate and sort suggestions; completion is invoked on every keystroke.
- **Tolerance**: real-world configs contain unknown keywords, malformed lines, missing includes, and unreadable files; ssh itself errors on unknown keywords, but completers typically skip and continue.
- **Over-approximation is acceptable**: static extraction cannot know whether a `Match` block applies; surfacing its declared hosts is usually preferable to missing them.
- **Documented limitations**: static facts are hints, not `ssh -G`-resolved configuration; say so in tool documentation where relevant.

## Sharing implementation

- Use one shared host-discovery implementation across tools to avoid divergent behavior between `ssh`/`scp`/`sftp` completers.
- Keep integrations narrow (depend on stable outputs) so extraction rules can evolve.

## Related

- [defaults.md](defaults.md) — `-F` and default discovery semantics.
- [host-candidates.md](host-candidates.md) — candidate extraction.
- [known-hosts.md](known-hosts.md) — known_hosts enrichment.
- [static-facts.md](static-facts.md) — per-host facts.
- [testing.md](testing.md) — safety and regression coverage.
