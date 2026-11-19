---
name: openssh
description: >
  Use when working with OpenSSH client configuration for shell completion: ssh_config parsing
  and indexing, Host/Match extraction, Match host/originalhost handling, Match exec/command
  safety, Include resolution, -F and default config discovery, host candidates, known_hosts,
  static per-host facts, and option/keyword metadata. Triggers on: "OpenSSH", "openssh",
  "ssh_config", "Host", "Match", "Match exec", "Match host", "Match originalhost", "Include",
  "-F", "known_hosts", "HostName", "User", "Port", "ProxyJump", "IdentityFile",
  "UserKnownHostsFile", "GlobalKnownHostsFile", "host candidates", "completion-safe".
user-invocable: true
---

# OpenSSH Completion Knowledge Base

Reference for implementing completion-safe, tolerant handling of OpenSSH client configuration. Focused on extracting useful completion data from `ssh_config`, `known_hosts`, and related sources without executing commands or requiring runtime connection evaluation.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

| Keywords | Reference |
|----------|----------|
| scope, goals, non-goals, completion-safe guarantees, tolerant parsing, deterministic results, separation of concerns | [references/overview.md](references/overview.md) |
| ssh_config syntax, tokenizer, parser, directives, Host, Match, Include, case-insensitive keywords, quoting, = separator, comments, whitespace, line continuation, tolerant parsing | [references/ssh-config-syntax.md](references/ssh-config-syntax.md) |
| Match, criteria, host, originalhost, exec, command, all, canonical, final, user, localnetwork, negation, comma-separated, token arity, no process execution, tolerant skipping | [references/match.md](references/match.md) |
| Include, glob, recursion, cycle detection, base dir, relative paths, ~ expansion, quoted paths, missing files, unreadable includes, including-file context | [references/include.md](references/include.md) |
| config sources, defaults, user config, system config, ~/.ssh/config, /etc/ssh/ssh_config, -F, custom path isolation, ssh_config.d, drop-ins | [references/defaults.md](references/defaults.md) |
| host candidates, declared hosts, Host arguments, Match host/originalhost extraction, wildcard filtering, negated names, dedupe, sort, deterministic output | [references/host-candidates.md](references/host-candidates.md) |
| known_hosts, hashed hosts, bracketed host:port, @cert-authority, @revoked, UserKnownHostsFile, GlobalKnownHostsFile, enrichment, merge, dedup, host sources | [references/known-hosts.md](references/known-hosts.md) |
| static facts, per-host directives, HostName, User, Port, IdentityFile, CertificateFile, ProxyJump, ProxyCommand, IdentityAgent, best-effort association, ssh -G limitations | [references/static-facts.md](references/static-facts.md) |
| option metadata, keywords, aliases, deprecated options, value types, enums, Match criterion metadata, argument arity, option/value completion | [references/option-metadata.md](references/option-metadata.md) |
| completer integration, host actions, ssh/scp/sftp consumers, -F flag propagation, contract stability, behavior regressions | [references/integration.md](references/integration.md) |
| testing, testdata fixtures, table-driven tests, safety tests, Match exec side effects, fuzzing, integration tests, regression cases, golden cases | [references/testing.md](references/testing.md) |

## Quick Guide

- **What guarantees should OpenSSH completion parsing provide?** → [references/overview.md](references/overview.md)
- **How should ssh_config be tokenized/parsed tolerantly?** → [references/ssh-config-syntax.md](references/ssh-config-syntax.md)
- **How do I extract hosts from `Match` without executing commands?** → [references/match.md](references/match.md) and [references/host-candidates.md](references/host-candidates.md)
- **How should `Include` resolution, globs, and cycles work?** → [references/include.md](references/include.md)
- **How do `-F`, user config, and system config interact?** → [references/defaults.md](references/defaults.md)
- **Which host tokens are completable and how are they filtered/deduped?** → [references/host-candidates.md](references/host-candidates.md)
- **How should `known_hosts` be parsed and merged?** → [references/known-hosts.md](references/known-hosts.md)
- **What per-host facts can be exposed without `ssh -G`?** → [references/static-facts.md](references/static-facts.md)
- **How should option/Match metadata be modeled?** → [references/option-metadata.md](references/option-metadata.md)
- **How should host/config logic integrate into completers?** → [references/integration.md](references/integration.md)
- **What tests prove completion-safety and tolerance?** → [references/testing.md](references/testing.md)

## Scope Notes

- This skill covers OpenSSH config and host-source semantics for completion.
- It does not replace OpenSSH's full runtime evaluator: `ssh -G` parity, canonicalization, token expansion, and dynamic Match evaluation are out of scope for default completion paths.
- Completion-safe behavior means never executing `Match exec`/`Match command` or any external process, and never requiring network access.
