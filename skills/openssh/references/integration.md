# carapace-bin Integration

How `carapace-ssh` is adopted by carapace-bin to fix host completion and provide a shared foundation for SSH-related completers (Phase 2 primary, later phases extend usage).

> **Source of truth**: [plan.md](../../../plan.md), Phase 2 and integration plan.

## Primary goal

Replace carapace-bin’s ad-hoc host scanner with a library API that fixes Match-related host loss ([carapace-bin#3935](https://github.com/carapace-sh/carapace-bin/pull/3935)) while keeping completion behavior stable where correct.

## Key integration points

### `pkg/actions/net/ssh/host.go`

- Replace implementation of `ActionHosts(config string)` to delegate to `carapace-ssh`.
- Preserve the action’s behavior/contract, tag, and filtering semantics as appropriate.
- `config` corresponds to the explicit config path (e.g. from `-F`) when provided; empty/default cases should use library default discovery as defined.

### Completers using host actions

- `ssh`, `scp`, `sftp` completers (e.g. `completers/common/{ssh,scp,sftp}_completer`) consume `ssh.ActionHosts(...)`.
- Audit `-F` flag propagation: ensure custom config path is passed correctly and defaults remain sensible.
- No completer-specific rewrites should be required beyond benefiting from the shared fix, if the action contract is preserved.

### Dependency cleanup

- Remove ad-hoc config parsing from carapace-bin once replaced.
- Remove `carapace-sh/ssh_config` / `kevinburke/ssh_config` dependency path and any related replace directives once unused.

## Behavioral contract

- **`-F` isolation**: explicit config path indexes only that entry (plus its includes), no implicit merge of user/system defaults. See [defaults.md](defaults.md).
- **Default behavior**: when no explicit config is provided, use documented default discovery (user/system entries, later drop-ins as defined). MVP must not regress #3935.
- **Completion-safe**: no process execution, no network access.
- **Deterministic output**: filtered, deduped, stably ordered candidates as appropriate. See [host-candidates.md](host-candidates.md).
- **Intentional improvements**: hosts previously dropped due to Match handling may now appear; document such improvements in release/integration notes as needed.

## Downstream and extended usage (later phases)

As `carapace-ssh` grows, carapace-bin can adopt:
- **known_hosts enrichment** for `ssh`/`scp`/`sftp` targets and tools like `ssh-keygen -F|-R` ([known-hosts.md](known-hosts.md)).
- **Static per-host facts** for host-aware identity/certificate/port/user/jump suggestions ([static-facts.md](static-facts.md)).
- **Option/keyword metadata** to drive/validate `ActionOptions` ([option-metadata.md](option-metadata.md)).
- **Destination/path helpers** for consistent `user@host`, `[host]:port`, scp/sftp/URI/jump positional completion (Phase 6+).

## Testing in carapace-bin

- Regression tests for `--macro net.ssh.Hosts(...)`.
- Completer-level behavior for positional host completion in `ssh`, `scp`, `sftp`.
- `-F` handling and default-path behavior.
- Cases from #3935: configs with `Match exec`, `Match originalhost`, and other criteria must still yield expected host candidates.

Library-side tests remain the primary guarantee; carapace-bin tests verify integration and contract stability. See [testing.md](testing.md).

## Migration notes

- Treat `carapace-ssh` as the canonical implementation for host discovery logic; avoid reintroducing parallel ad-hoc parsers.
- Keep API usage narrow in carapace-bin to ease future library evolution (additive changes, clear versioning).
- If default drop-in discovery rules change, ensure behavior is deliberate, configurable, and well-tested to avoid surprises across environments.

## Related

- [overview.md](overview.md) — phases and definition of done.
- [defaults.md](defaults.md) — source modes and `-F` isolation.
- [testing.md](testing.md) — regression and safety coverage.
