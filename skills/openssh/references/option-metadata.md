# Keyword and Match Criterion Metadata

Reference facts for ssh_config keywords and Match criteria that help parsing and option/value completion.

## Keyword facts

- Keywords are **case-insensitive**; arguments are case-sensitive (see [ssh-config-syntax.md](ssh-config-syntax.md)).
- Directives accept whitespace or a single `=` as separator.
- Unknown keywords are an error for ssh unless listed in `IgnoreUnknown`.
- Some keywords are repeatable and additive (`IdentityFile`, `CertificateFile`, forwardings); most follow first-value-wins.
- Deprecated aliases exist (e.g. `ChallengeResponseAuthentication` is an alias of `KbdInteractiveAuthentication`; `HostbasedKeyTypes` was renamed to `HostbasedAcceptedAlgorithms`).
- Argument types encountered across keywords: string, path (with `~`/TOKENS/env expansion per directive), integer/port, yes/no or multi-state booleans (e.g. `yes|no|ask|confirm`), enumerations (`AddressFamily`: `any|inet|inet6`), pattern-lists, time intervals, and command strings (`ProxyCommand`, `LocalCommand`, `RemoteCommand`, `KnownHostsCommand`).

## Match criterion arity

| Argument shape | Criteria |
|----------------|----------|
| zero-argument | `all`, `canonical`, `final` |
| single value / comma-separated pattern-list | `host`, `originalhost`, `user`, `localuser`, `tagged`, `command`, `version` |
| specific forms | `localnetwork` (CIDR list), `sessiontype` (one of `shell`, `exec`, `subsystem`, `none`) |
| free-form / command | `exec` (command text; runs under the user's shell — **never execute**) |

`all` must appear alone or immediately after `canonical`/`final`; other criteria combine arbitrarily; any criterion may be negated with `!`.

sshd_config(5) uses a different criteria set; keep client and server metadata separate.

## Notes for maintainers

- The keyword/criteria sets evolve across OpenSSH releases; new keywords appear regularly and old ones get renamed. Tolerance-first handling (ignore unknowns) avoids breakage.
- Version notes worth recording: `tagged`, `sessiontype`, `version`, `localnetwork`, `command` are recent additions; consult ssh_config(5) for the version being targeted.

## Related

- [match.md](match.md) — Match criteria semantics.
- [ssh-config-syntax.md](ssh-config-syntax.md) — syntax and PATTERNS/TOKENS.
- [static-facts.md](static-facts.md) — directives worth extracting.
