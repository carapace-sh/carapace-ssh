# Config Sources and Default Discovery

How to model explicit config paths (`-F`) versus default user/system configuration, and how those choices affect include resolution and host results.

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

Default discovery should provide sensible, well-defined behavior and can include drop-in defaults when rules are clearly specified and tested.

## API shape (illustrative)

Implementations commonly separate source/context description from indexing and querying, for example:

```go
// Context for resolution (illustrative)
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

Useful options include custom base directories, include resolver overrides for tests, and read/symlink policies where needed. Keep the public surface small and stable for consumers.

## Drop-in and distro-specific includes

Common real-world setups include drop-in directories, for example:
- `/etc/ssh/ssh_config.d/*.conf`
- user-side `config.d` patterns (distro/build dependent)

Guidance:
- include drop-ins only with **clearly defined, tested rules** aligned with expected OpenSSH behavior
- **do not guess destructively** (avoid surprising behavior changes)
- make default discovery **configurable** where variation is expected
- document the delivered search behavior

## Interaction with completers

For SSH-related completers, host/config discovery should make it easy to:
- index the exact file specified by `-F` (isolated)
- index discovered defaults when no `-F` is provided
- keep behavior consistent across tools that share a common host-discovery contract

See [integration.md](integration.md).

## Determinism and safety

- No hidden global state: resolution depends on explicit inputs/options.
- No network access, no process execution.
- Deterministic, deduplicated results where appropriate (see [host-candidates.md](host-candidates.md)).

## Testing expectations

- **Custom `-F` isolation**: hosts/includes from user/system defaults must not leak when an explicit path is indexed, except what that file (and its includes) provides.
- Default-path behavior tested against controlled fixtures/temp dirs (resolver overrides helpful for tests).

See [testing.md](testing.md).

## Related

- [include.md](include.md) — base-dir and path resolution during include traversal.
- [host-candidates.md](host-candidates.md) — candidate extraction/filtering outputs.
- [integration.md](integration.md) — completer contracts and `-F` propagation.
