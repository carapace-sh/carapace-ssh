# Configuration File Discovery

Which files ssh(1) reads and how `-F`, default discovery, and drop-in directories interact. Facts per ssh(1)/ssh_config(5), plus commonly observed distro conventions.

## ssh(1) defaults

- User configuration: `~/.ssh/config`
- System-wide configuration: `/etc/ssh/ssh_config`
- Precedence: command-line options, then user config, then system config; **first obtained value for each directive wins**.

## `ssh -F configfile`

- Specifies an alternative per-user configuration file.
- When a configuration file is given on the command line, **the system-wide configuration file is ignored**.
- The given file's `Include` directives still apply (resolved relative to the user context, i.e. `~/.ssh`).

For completion tools that mirror ssh, `-F` semantics translate to: index only the explicitly given file (plus its includes) and do not merge in default user/system configs.

## `ssh -G host`

- Prints the effective configuration after evaluating all sources for a host; the authoritative resolution. May canonicalize the hostname (depending on `CanonicalizeHostname`) and may therefore perform DNS lookups.

## Drop-in directories (distro conventions)

Common real-world setups use drop-in directories included from the main config, for example:

- `/etc/ssh/ssh_config.d/*.conf` — included by an `Include` line in `/etc/ssh/ssh_config` on many distributions (Debian, Fedora, ...).
- User-side `~/.ssh/config.d/` patterns exist but are **not** part of OpenSSH itself.

A completion tool that wants drop-in support should treat it as following `Include` from the entry file (which is what ssh does), not as an implicit extra source.

## Base-directory contexts

- Relative `Include` paths resolve to `~/.ssh` when included from a user configuration file, and `/etc/ssh` when included from the system configuration file (see [include.md](include.md)).
- `~` expands to the local user's home directory; TOKENS and environment variables expand at connection time only.

## Defaults of commonly referenced paths

| Directive | Default |
|-----------|---------|
| `GlobalKnownHostsFile` | `/etc/ssh/ssh_known_hosts`, `/etc/ssh/ssh_known_hosts2` |
| `UserKnownHostsFile` | `~/.ssh/known_hosts`, `~/.ssh/known_hosts2` |
| `IdentityFile` | `~/.ssh/id_rsa`, `~/.ssh/id_ecdsa`, `~/.ssh/id_ecdsa_sk`, `~/.ssh/id_ed25519`, `~/.ssh/id_ed25519_sk`, `~/.ssh/id_mldsa44_ed25519` |
| `Port` | 22 |

## Related

- [include.md](include.md) — Include path resolution.
- [overview.md](overview.md) — source order and precedence.
- [known-hosts.md](known-hosts.md) — known_hosts file locations.
