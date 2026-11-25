---
name: openssh
description: >
  Use when working with OpenSSH client configuration and related data: ssh_config syntax
  and semantics (per ssh_config(5)), Host/Match blocks, PATTERNS and TOKENS, Match criteria
  (canonical, final, exec, host, originalhost, user, localuser, localnetwork, tagged,
  command, version, sessiontype), Include directives, configuration file discovery and
  precedence (-F, -G, user/system configs, ssh_config.d), known_hosts file format
  (markers, bracketed host:port, hashed entries), host-specific directives (HostName, User,
  Port, ProxyJump, IdentityFile, UserKnownHostsFile), and considerations for implementing
  shell completion over these sources. Triggers on: "OpenSSH", "openssh", "ssh_config",
  "Host", "Match", "Include", "-F", "-G", "ssh -G", "known_hosts", "PATTERNS", "TOKENS",
  "HostName", "User", "Port", "ProxyJump", "IdentityFile", "UserKnownHostsFile",
  "GlobalKnownHostsFile", "ssh-keygen -F", "ssh-keygen -R", "completion-safe".
user-invocable: true
---

# OpenSSH Client Configuration Knowledge Base

Factual reference for OpenSSH client configuration and host-key data, aimed at tools that extract completion information from `ssh_config` and `known_hosts` without evaluating runtime conditions.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

| Keywords | Reference |
|----------|----------|
| configuration sources and precedence, command line / user / system, first-value-wins, `-F`, `-G`, ssh_config.d drop-ins | [references/overview.md](references/overview.md) and [references/defaults.md](references/defaults.md) |
| ssh_config file syntax, directive/value pairs, case-insensitivity, `=` separator, quoting, comments, PATTERNS, TOKENS, IgnoreUnknown | [references/ssh-config-syntax.md](references/ssh-config-syntax.md) |
| Match blocks, criteria list and argument arity, exec/host/originalhost/user/tagged/command/version/sessiontype, negation, canonical/final re-parse passes | [references/match.md](references/match.md) |
| Include directive, glob order, relative path contexts (`~/.ssh` vs `/etc/ssh`), conditional inclusion, tokens in paths | [references/include.md](references/include.md) |
| `Host` block arguments, `Match host`/`originalhost` values, pattern vs literal tokens, wildcards and negation, candidate filtering | [references/host-candidates.md](references/host-candidates.md) |
| known_hosts file format, markers `@cert-authority`/`@revoked`, bracketed `[host]:port`, hashed `|1|` entries, default file locations, ssh-keygen `-F`/`-R` | [references/known-hosts.md](references/known-hosts.md) |
| host-specific directives (HostName, User, Port, IdentityFile, CertificateFile, ProxyJump, ProxyCommand, IdentityAgent, UserKnownHostsFile, GlobalKnownHostsFile) and their defaults | [references/static-facts.md](references/static-facts.md) |
| keyword vocabulary, argument types, deprecated aliases, Match criterion arity metadata | [references/option-metadata.md](references/option-metadata.md) |
| wiring OpenSSH-derived data into completers, `-F` propagation, behavioral expectations | [references/integration.md](references/integration.md) |
| testing approaches for tolerant extraction, safety (never executing commands), regression coverage | [references/testing.md](references/testing.md) |

## Quick Guide

- **Where does ssh(1) read configuration from, and which value wins?** → [references/overview.md](references/overview.md) and [references/defaults.md](references/defaults.md)
- **What syntax does ssh_config allow (quotes, `=`, case, comments)?** → [references/ssh-config-syntax.md](references/ssh-config-syntax.md)
- **What Match criteria exist and what do they consume?** → [references/match.md](references/match.md)
- **How does `Include` resolve relative paths and globs?** → [references/include.md](references/include.md)
- **Which host tokens can be offered as completions?** → [references/host-candidates.md](references/host-candidates.md)
- **What is the known_hosts format and which entries are usable?** → [references/known-hosts.md](references/known-hosts.md)
- **Which directives declare per-host facts and what are their defaults?** → [references/static-facts.md](references/static-facts.md)
- **How do completers use this data safely?** → [references/integration.md](references/integration.md)

## Scope Notes

- This skill describes OpenSSH itself (man-page-level facts) plus general considerations for extracting completion data from it.
- It does not describe any specific library's API or internal design decisions.
- Anything requiring runtime evaluation (effective configuration) belongs to `ssh -G`, which is an external evaluator and may touch the network; static extraction must not execute commands or access the network.
