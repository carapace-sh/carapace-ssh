# Completer Integration

Guidance for integrating completion-safe OpenSSH config/host data into shell completers and related tools.

## Integration goals

- Provide consistent host suggestions for `ssh`, `scp`, `sftp`, and other tools that accept SSH destinations.
- Centralize config parsing, include resolution, and host-source merging to avoid parallel ad-hoc parsers with divergent behavior.
- Preserve a stable, documented contract for consumers (inputs, defaults, isolation, output shape).

## Common integration points

### Host suggestion action/helper

A typical integration exposes a host-suggestion function that:
- accepts an optional explicit config path (for `-F`-style behavior)
- returns filtered, completable host candidates
- uses deterministic ordering and deduplication
- remains completion-safe (no process execution, no network access)

Consumers (destination completers, option/value completers, macros) should delegate host discovery to this shared logic rather than reimplementing ssh_config scanning.

### Completers that consume host data

- Destination/positional completers for `ssh`, `scp`, `sftp`.
- Jump/proxy target completers (for `-J` / `ProxyJump` contexts) where host suggestions are appropriate.
- Tools that operate on known hosts or configured hosts, when they share the same discovery semantics.

Ensure flags that select an explicit config file are propagated to host discovery consistently.

## Behavioral contract

- **Explicit config isolation**: when a specific config file is provided, index only that entry (plus its includes) unless the caller explicitly requests merged/default behavior. Do not implicitly leak user/system defaults.
- **Default discovery**: when no explicit config is provided, use clearly defined and configurable default entry points (for example user and system configs, and any supported drop-ins).
- **Completion-safe**: never execute `Match exec`/`Match command` or any external process; no network access.
- **Deterministic output**: filtered, deduplicated, and stably ordered candidates as appropriate for UI/completion stability.
- **Tolerant behavior**: unreadable includes, unknown directives/criteria, or malformed lines should not prevent returning recoverable results from the rest of the config graph.
- **Intentional improvements**: better extraction from previously unsupported or poorly handled configs (for example hosts declared around non-evaluable `Match` criteria) may change suggestions; document such behavior as compatibility improvements when relevant.

## Extended integration patterns

As additional data sources become available, consumers may adopt:

- **known_hosts enrichment** for previously seen targets (with hashed-host limitations understood).
- **Static per-host facts** for host-aware identity/certificate, user, port, or jump suggestions (best-effort hints, not `ssh -G` resolution).
- **Option/keyword and Match metadata** to drive or validate option/value completion.
- **Destination/path helpers** for consistent handling of `user@host`, `[host]:port`, scp/sftp remote forms, URIs, and jump specifications.

## Testing considerations

- Verify host suggestions for explicit config paths and default discovery modes.
- Cover `-F` propagation and isolation in consumer-facing behavior.
- Include regression cases where hosts appear in configs with non-evaluable `Match` criteria (`exec`, `command`, `user`, `originalhost`, etc.).
- Assert no side effects during completion (especially for `Match exec`/`Match command`).
- Keep source-level parsing tests as the primary guarantee; integration tests should confirm contract stability and wiring.

## Migration and consistency notes

- Treat a shared implementation as the canonical host-discovery logic to prevent behavior drift.
- Keep consumer integrations narrow (depend on documented APIs/contracts) to ease evolution.
- When changing default discovery or filtering rules, make changes deliberate, configurable where needed, and well-tested to avoid surprising environment-specific regressions.

## Related

- [defaults.md](defaults.md) — source modes and explicit-config isolation.
- [host-candidates.md](host-candidates.md) — candidate extraction and filtering.
- [known-hosts.md](known-hosts.md) — host-source enrichment.
- [static-facts.md](static-facts.md) — best-effort directive hints.
- [option-metadata.md](option-metadata.md) — option and Match metadata reuse.
- [testing.md](testing.md) — regression and safety coverage.
