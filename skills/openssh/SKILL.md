---
name: openssh
description: >
  Use when working with OpenSSH client config for completion, especially carapace-ssh and
  plan.md implementation: ssh_config parsing/indexing, Host/Match extraction, Match host and
  originalhost, Match exec/command safety, Include resolution, -F and default config discovery,
  host candidates, known_hosts, static per-host facts, option/keyword metadata, carapace-bin
  integration. Triggers on: "OpenSSH", "openssh", "ssh_config", "Host", "Match", "Match exec",
  "Match host", "Match originalhost", "Include", "-F", "known_hosts", "HostName", "User",
  "Port", "ProxyJump", "IdentityFile", "UserKnownHostsFile", "GlobalKnownHostsFile",
  "carapace-ssh", "host candidates", "completion-safe".
user-invocable: true
---

# OpenSSH Completion Knowledge Base

In-depth reference for implementing `carapace-ssh`: a completion-safe, tolerant OpenSSH client configuration library for carapace shell completion. Grounded in [plan.md](../../plan.md) and the implementation constraints for `ssh`, `scp`, `sftp`, and related completers.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

| Keywords | Reference |
|----------|----------|
| project scope, goals, non-goals, phases, roadmap, MIT, fresh implementation, carapace-ssh, #3935, completion-safe guarantees | [references/overview.md](references/overview.md) |
| ssh_config syntax, tokenizer, parser, directives, Host, Match, Include, case-insensitive keywords, quoting, = separator, comments, whitespace, line continuation, tolerant parsing | [references/ssh-config-syntax.md](references/ssh-config-syntax.md) |
| Match, criteria, host, originalhost, exec, command, all, canonical, final, user, localnetwork, negation, comma-separated, token arity, no process execution, tolerant skipping | [references/match.md](references/match.md) |
| Include, glob, recursion, cycle detection, base dir, relative paths, ~ expansion, quoted paths, missing files, unreadable includes, including-file context | [references/include.md](references/include.md) |
| config sources, defaults, user config, system config, ~/.ssh/config, /etc/ssh/ssh_config, -F, custom path isolation, ssh_config.d, drop-ins, Source, IndexFile, IndexDefault | [references/defaults.md](references/defaults.md) |
| host candidates, declared hosts, Host arguments, Match host/originalhost extraction, wildcard filtering, negated names, dedupe, sort, deterministic output, HostCandidates, DeclaredHosts | [references/host-candidates.md](references/host-candidates.md) |
| known_hosts, hashed hosts, bracketed host:port, @cert-authority, @revoked, UserKnownHostsFile, GlobalKnownHostsFile, enrichment, merge, dedup, host sources | [references/known-hosts.md](references/known-hosts.md) |
| static facts, per-host directives, HostName, User, Port, IdentityFile, CertificateFile, ProxyJump, ProxyCommand, best-effort association, ssh -G limitations | [references/static-facts.md](references/static-facts.md) |
| option metadata, keywords, aliases, deprecated options, value types, enums, Match criterion metadata, argument arity, ActionOptions, config/CLI knowledge | [references/option-metadata.md](references/option-metadata.md) |
| carapace-bin integration, ActionHosts, pkg/actions/net/ssh/host.go, ssh/scp/sftp completers, -F flag propagation, regression behavior, remove ssh_config dependency | [references/integration.md](references/integration.md) |
| testing, testdata fixtures, table-driven tests, safety tests, Match exec side effects, fuzzing, integration tests, #3935 regressions, golden cases | [references/testing.md](references/testing.md) |

## Quick Guide

- **What is `carapace-ssh` and what must it guarantee?** → [references/overview.md](references/overview.md)
- **How should ssh_config be tokenized/parsed tolerantly?** → [references/ssh-config-syntax.md](references/ssh-config-syntax.md)
- **How do I extract hosts from `Match` without executing commands?** → [references/match.md](references/match.md) and [references/host-candidates.md](references/host-candidates.md)
- **How should `Include` resolution, globs, and cycles work?** → [references/include.md](references/include.md)
- **How do `-F`, user config, and system config interact?** → [references/defaults.md](references/defaults.md)
- **Which host tokens are completable and how are they filtered/deduped?** → [references/host-candidates.md](references/host-candidates.md)
- **How should `known_hosts` be parsed and merged?** → [references/known-hosts.md](references/known-hosts.md)
- **What per-host facts can be exposed without `ssh -G`?** → [references/static-facts.md](references/static-facts.md)
- **How should option/Match metadata be modeled for reuse?** → [references/option-metadata.md](references/option-metadata.md)
- **How will this replace carapace-bin's host action?** → [references/integration.md](references/integration.md)
- **What tests prove completion-safety and fix #3935?** → [references/testing.md](references/testing.md)
- **What are the implementation phases and MVP definition of done?** → [references/overview.md](references/overview.md)

## Cross-Project References

- For carapace completion action design and integration patterns in Go completers, use the **carapace** and **carapace-dev** skills.
- For shell-facing completion behavior and integration context, consult the relevant shell skills (for example **bash**, **zsh**, **fish**) when implementing or debugging consumer-facing behavior.
- This skill covers OpenSSH config/host-source semantics for `carapace-ssh`; it does not replace OpenSSH's full runtime evaluator (`ssh -G` parity is explicitly out of scope for default completion paths).
