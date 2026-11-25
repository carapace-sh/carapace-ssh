# OpenSSH Client Configuration Overview

Man-page-level facts about how `ssh(1)` obtains and applies client configuration, and the constraints that matter for tools extracting completion data from it.

## Configuration sources and precedence

`ssh(1)` obtains configuration data from the following sources, in order:

1. command-line options
2. the user's configuration file (`~/.ssh/config`)
3. the system-wide configuration file (`/etc/ssh/ssh_config`)

Unless noted otherwise, **the first obtained value for each directive is used**; subsequent values are ignored. Consequently, host-specific declarations should appear near the beginning of the files and general defaults at the end.

Files are organized in sections introduced by conditional `Host` or `Match` directives. Configuration inside a section is applied only if that section's condition matches.

- `ssh -F configfile` selects an alternative per-user configuration file; when given, the system-wide configuration file is ignored.
- `ssh -G host` prints the fully evaluated effective configuration for a host (after `Host`/`Match` evaluation and hostname canonicalization). This is the only way to obtain a complete resolution, and it runs external code only in the sense of the ssh binary itself (canonicalization may perform DNS lookups).

## Structure of the data sources

- `ssh_config` files: directive/value pairs, sections via `Host`/`Match`, `Include` for composition. See [ssh-config-syntax.md](ssh-config-syntax.md), [match.md](match.md), [include.md](include.md).
- `known_hosts` files: host-key database consulted for host authentication; also a source of previously seen host names. See [known-hosts.md](known-hosts.md).
- Host-specific directives (`HostName`, `User`, `Port`, `IdentityFile`, ...) declare per-host facts. See [static-facts.md](static-facts.md).

## Runtime evaluation is not statically available

Several mechanisms only have meaning at connection time and must not be simulated during static extraction:

- `Match exec` runs a command under the user's shell and checks its exit status.
- `KnownHostsCommand`, `ProxyCommand`, `LocalCommand` execute commands; treat their argument strings as opaque data.
- TOKENS (`%h`, `%r`, `%p`, `%C`, ...) and environment variables are substituted at connection time.
- `CanonicalizeHostname` can cause the configuration to be re-parsed with a canonical name; the effective result depends on DNS.
- `ssh -G` is the authoritative evaluator.

## Completion-safety considerations

Tools that extract completion data from these sources commonly aim to:

- never execute `Match exec` or any directive whose value is a command string;
- never access the network;
- treat unknown keywords, malformed lines, and unreadable files tolerantly rather than aborting (real-world configs contain all of these; ssh itself errors on unknown keywords unless `IgnoreUnknown` is set — see [ssh-config-syntax.md](ssh-config-syntax.md));
- return deduplicated, deterministically ordered results.

## Related

- [ssh-config-syntax.md](ssh-config-syntax.md) — file syntax, PATTERNS, TOKENS.
- [match.md](match.md) — Match criteria and re-parse passes.
- [include.md](include.md) — Include directive semantics.
- [defaults.md](defaults.md) — source selection (`-F`, user/system, drop-ins).
- [host-candidates.md](host-candidates.md) — host tokens and completable filtering.
- [known-hosts.md](known-hosts.md) — known_hosts format.
- [static-facts.md](static-facts.md) — host-specific directives.
- [option-metadata.md](option-metadata.md) — keyword/criterion metadata.
- [integration.md](integration.md) — use in completers.
- [testing.md](testing.md) — testing approaches.
