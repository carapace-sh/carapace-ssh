# ssh_config File Syntax

Syntax of OpenSSH client configuration files per ssh_config(5).

## File structure

- The file contains directive/value pairs, one per line.
- Lines starting with `#` and empty lines are comments/ignored.
- `#` outside of a quoted string may be used to add a comment to the end of a line.
- Whitespace at the beginning or end of lines is ignored.
- **Keywords are case-insensitive; arguments are case-sensitive.**
- A directive is separated from its value by whitespace, or by **exactly one `=` character** (which may be surrounded by whitespace). The `=` form exists so options passed via `ssh -o` need not quote whitespace.
- Values may optionally be enclosed in double quotes (`"`) to represent arguments containing spaces.
- OpenSSH itself does not support backslash line continuations; real-world files sometimes contain them, so tolerant parsers may treat them best-effort.

## Host directive

`Host` restricts the following declarations (up to the next `Host` or `Match`) to hosts matching one of the patterns given after the keyword:

- Multiple patterns are separated by whitespace.
- Patterns use glob syntax: `*` matches any sequence of characters, `?` matches any single character.
- A pattern prefixed with `!` is negated; **if a negated entry matches, the Host entry is ignored regardless of any other patterns on the line** (negations act as exceptions to wildcard matches).
- A single `*` provides global defaults.
- The host matched is usually the hostname given on the command line (with exceptions when `CanonicalizeHostname` is enabled).

## Match directive

See [match.md](match.md). `Match` introduces a conditional section; conditions are criteria on the `Match` line itself, not separate directives.

## Include directive

See [include.md](include.md).

## Other keywords

OpenSSH supports many keywords (a few dozen, from `AddKeysToAgent` to `UpdateHostKeys`); they carry values of types such as strings, paths, integers/ports, yes/no booleans, enumerations, and command strings. Examples of completion-relevant ones: `HostName`, `User`, `Port`, `IdentityFile`, `CertificateFile`, `ProxyJump`, `ProxyCommand`, `IdentityAgent`, `UserKnownHostsFile`, `GlobalKnownHostsFile` (see [static-facts.md](static-facts.md)).

Unknown keywords cause ssh to fail with an error, unless listed in `IgnoreUnknown` (a comma-separated pattern list of option names to ignore). Tolerant completion parsers typically ignore unknown keywords instead.

## PATTERNS (per ssh_config(5))

- A pattern is zero or more characters followed by `*` (matches any number of characters), `?` (matches exactly one character), or ordinary characters.
- A pattern may be negated by prefixing it with `!`.
- Several patterns may be separated by commas to form a pattern-list; matching succeeds if any (non-negated) pattern matches and no negated pattern matches.
- Wildcards and negation appear in `Host` arguments, `Match` criteria values, `known_hosts` host fields, and various pattern-list options (e.g. `SendEnv`).

## TOKENS (per ssh_config(5))

Arguments of some directives (those that mention "accept the tokens") expand the following before use; they are only meaningful at connection time:

| Token | Meaning |
|-------|---------|
| `%%` | literal `%` |
| `%C` | hash of `%l%h%p%r%j` |
| `%d` | local user's home directory |
| `%f` | fingerprint of the server host key |
| `%h` | remote hostname |
| `%I` | key identity of the server host key |
| `%i` | local user's uid |
| `%j` | jump host (empty if none) |
| `%K` | base64-encoded server host key |
| `%k` | server host key |
| `%L` / `%l` | local hostname |
| `%n` | original command-line hostname |
| `%p` | remote port |
| `%r` | remote username |
| `%T` / `%t` | target tun device / local tun device |
| `%u` | local username |

Static extraction should preserve these tokens as written rather than expanding them (their values depend on the connection context).

## Related

- [match.md](match.md) — Match criteria grammar.
- [include.md](include.md) — Include semantics.
- [host-candidates.md](host-candidates.md) — which host tokens become completions.
- [overview.md](overview.md) — sources, precedence, safety.
