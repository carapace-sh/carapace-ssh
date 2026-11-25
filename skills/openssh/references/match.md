# Match Directive

Semantics of `Match` blocks per ssh_config(5), and what is relevant when extracting data from them statically.

## Definition

`Match` restricts the following declarations (up to the next `Host` or `Match`) to when the conditions following the keyword are satisfied. Conditions are specified using one or more criteria, or the single token `all` which always matches.

**The available criteria are: `canonical`, `final`, `exec`, `localnetwork`, `host`, `originalhost`, `tagged`, `command`, `user`, `localuser`, `version`, and `sessiontype`.**

- `all` must appear alone or immediately after `canonical` or `final`.
- Other criteria may be combined arbitrarily.
- **All criteria except `all`, `canonical`, and `final` require an argument.**
- Criteria may be negated by prepending `!` (e.g. `Match !user root`).
- Criteria accept the same separator forms as directives (`Match host=example.com` and `Match host example.com`).

## Criteria reference

| Criterion | Argument | Semantics |
|-----------|----------|-----------|
| `canonical` | none | matches only during the re-parse after hostname canonicalization (`CanonicalizeHostname`) |
| `final` | none | requests a final re-parse of the configuration and matches only during that pass; with `CanonicalizeHostname` enabled, `canonical` and `final` match during the same pass |
| `exec` | command text (quoted if it contains spaces) | **executes the command under the user's shell** (`SHELL` from the environment); zero exit status = condition true; arguments accept TOKENS |
| `localnetwork` | CIDR network list | matches addresses of active local network interfaces |
| `host` | pattern-list (single entry or comma-separated) | matches the target hostname, **after** substitution by `Hostname` or `CanonicalizeHostname` |
| `originalhost` | pattern-list | matches the hostname as specified on the command line |
| `tagged` | pattern-list | matches a tag from a prior `Tag` directive or `ssh -P`; the empty string matches the case where no tag is specified (`Match tagged ""`) |
| `command` | pattern-list | matches the requested remote command or subsystem name (e.g. `sftp`); the empty string matches the case where no command was requested |
| `user` | pattern-list | matches the target username on the remote host |
| `localuser` | pattern-list | matches the name of the local user running ssh (useful in system-wide configs) |
| `version` | pattern-list | matches the ssh client version string, e.g. `OpenSSH_10.0` |
| `sessiontype` | one of `shell`, `exec`, `subsystem`, `none` | matches the requested session type (`-N` corresponds to `none`) |

Pattern-lists use the PATTERNS operators (wildcards `*`/`?` and negation `!`); see [ssh-config-syntax.md](ssh-config-syntax.md).

Note: sshd_config(5) has a **different** Match criteria set (`user`, `group`, `host`, `address`, `localaddress`, `localport`, `ruser`, `rdomain`, `all`); do not mix the two lists up when working with client configuration.

## OpenSSH versions

Criteria have been added over time (`tagged`, `sessiontype`, `version`, `command`, `localnetwork` are among the newest). Parsers should treat unknown criteria tolerantly rather than erroring.

## Static extraction considerations

- `exec` runs arbitrary commands; extraction must never execute it. Its argument extends to the criteria's value (quoted text) — treat it as opaque data. Note that because `exec` accepts free-form command text, token alignment for criteria after it on the same line is inherently ambiguous; conservative parsers stop extracting on that criterion.
- `host` and `originalhost` values are the natural source of host candidates: extract their pattern-lists unless the criterion (or individual value) is negated.
- Zero-argument criteria (`all`, `canonical`, `final`) consume nothing; value criteria consume one argument each (with `criterion=value` and spaced forms).
- Whether a `Match` block would actually apply at connection time is not statically decidable in general; extraction of declared tokens is a best-effort over-approximation, not a resolution.

## Examples

```sshconfig
Match all
    # always applies

Match canonical all
    # applies during canonicalized re-parse

Match originalhost bastion,*.example.com
    HostName 10.0.0.1

Match host prod-*,!prod-db user alice
    User deploy

Match exec "test -f /tmp/marker"
    # executes a shell command; never do this during extraction

Match tagged "" user deploy
    # empty tag means "no tag specified"

Match sessiontype none
    # transport-only sessions (-N)
```

## Related

- [ssh-config-syntax.md](ssh-config-syntax.md) — PATTERNS and TOKENS.
- [host-candidates.md](host-candidates.md) — host token extraction and filtering.
- [option-metadata.md](option-metadata.md) — criterion arity metadata.
- [overview.md](overview.md) — sources, precedence, safety.
