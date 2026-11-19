# Config Sources and Default Discovery

How `carapace-ssh` should model explicit config paths (`-F`) versus default user/system configuration, and how those choices affect include resolution and host results.

> **Source of truth**: [plan.md](../../../plan.md), Phase 1 and Phase 2.

## Design principle: explicit context

Callers control/provide config path(s), base directories, and system/user roots; no hidden global state. Defaults are opt-in helpers. This keeps behavior testable and avoids surprising completion results.

## Entry modes

### Explicit config path (`-F` / custom path)

When the caller specifies an explicit config file (as `ssh -F` / completer `-F` implies):

- Read **only that file as the entry**.
- **Do not implicitly merge** system or user default configs.
- **Do follow its `Include` directives** (resolved according to the documented/configured context for that source).

This isolation is important for correctness and is called out as a required test case.

### Default discovery

Helpers may index standard entry points:

| Context | Typical entry | Typical base/intent |
|---------|---------------|---------------------|
| User | `~/.ssh/config` | User SSH directory context (e.g. `~/.ssh`) |
| System | `/etc/ssh/ssh_config` | System SSH directory context (e.g. `/etc/ssh`) |

MVP must not regress #3935 behavior while providing sensible defaults. Phase 2 can expand drop-in defaults with clearly defined, tested rules.

## API sketch (initial)

From the plan (subject to refinement):

```go
// Context for resolution (illustrative; exact types TBD)
type Source struct {
    Path    string
    BaseDir string
    System  bool
}

func IndexFile(path string, opts ...Option) (*Index, error)
// IndexPaths(...), IndexDefault(...)

func (idx *Index) HostCandidates() []string
// and/or DeclaredHosts() for richer info
```

Possible options: custom base dir, include resolver overrides (for tests), read limits/symlink policy if needed.

Public API should remain small and stable; expose only what carapace-bin and future consumers need.

## Drop-in and distro-specific includes (Phase 2)

Common real-world setups include drop-in directories, for example:
- `/etc/ssh/ssh_config.d/*.conf`
- user-side `config.d` patterns (distro/build dependent)

Phase 2 guidance:
- consider only with **clearly defined, tested rules** aligned with OpenSSH behavior expectations
- **do not guess destructively** (avoid surprising behavior changes)
- make defaults/config discovery **configurable** where variation is expected
- document default search behavior as delivered

## Interaction with completers

In carapace-bin, `ssh`, `scp`, and `sftp` completers pass/interpret `-F` for host completion. The library must make it easy to:
- index the exact file specified by `-F` (isolated)
- index discovered defaults when no `-F` is provided
- keep behavior consistent across tools that share `ActionHosts(config string)`

See [integration.md](integration.md).

## Determinism and safety

- No hidden global state: resolution depends on explicit inputs/options.
- No network access, no process execution.
- Deterministic, deduplicated results where appropriate (see [host-candidates.md](host-candidates.md)).

## Testing expectations

- **Custom `-F` isolation**: hosts/includes from user/system defaults must not leak when an explicit path is indexed, except what that file (and its includes) provides.
- Default-path behavior tested against controlled fixtures/temp dirs (resolver overrides helpful for tests).
- Regression coverage for #3935 scenarios under both explicit and default contexts as applicable.

See [testing.md](testing.md).

## Related

- [include.md](include.md) — base-dir and path resolution during include traversal.
- [host-candidates.md](host-candidates.md) — candidate extraction/filtering outputs.
- [integration.md](integration.md) — carapace-bin `-F` and `ActionHosts` contract.
