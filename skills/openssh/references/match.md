# Match Handling (Completion-Safe)

How to parse `Match` directives for host-candidate extraction without evaluating runtime conditions or executing commands.

## Core rule: never execute

The indexer/parser **must never execute**:
- `Match exec ...`
- `Match command ...`
- any external process or shell command implied by Match criteria

It also must not require runtime connection context (current user, network, resolved hostname, canonicalization state, etc.). Parsing is static and completion-safe by default.

## What to extract

Collect literal host candidates from `Match` criteria:
- `host` — match against the target hostname as given
- `originalhost` — match against the original hostname as specified on the command line

Include comma-separated lists and preserve awareness of `!`-negated forms as they appear in criteria arguments (see [host-candidates.md](host-candidates.md) for how negations/wildcards affect the *completable* filtered set). Do **not** extract hosts from negated `host`/`originalhost` criteria (e.g. `!host bad.example`), though their arguments must still be consumed for alignment. Extraction is static/best-effort and must not rely on runtime evaluation.

## Behavioral requirements (tolerant, non-executing)

| Requirement | Detail |
|-------------|--------|
| **No execution** | Never run `exec`/`command` or any external process. Treat `exec` command arguments as opaque (never executed); `command` takes a command/subsystem pattern argument (possibly empty/quoted) and must also never be executed. |
| **Zero-arg criteria** | Parse/skip known zero-arg criteria such as `all`, `canonical`, `final` (and `canonical all` forms as encountered). |
| **Argument-consuming criteria** | Parse/skip criteria that consume arguments so token alignment is preserved well enough to still find `host`/`originalhost` in the same `Match` line. |
| **Unknown criteria/keywords** | Must not abort parsing of the file or prevent host extraction elsewhere. Fall back safely. |
| **Host extraction survives gating** | Hosts declared in blocks gated by non-evaluable/unsupported criteria must still be discoverable as declared candidates where applicable to extraction goals (especially `Match host|originalhost` values themselves). |

## Known criteria to account for

Tolerant handling commonly needs to account for criteria including:
- `exec` (command argument; **never execute**, opaque; may prevent safe recovery of later criteria on the same line)
- `command` (requested command/subsystem pattern argument; **never execute**; may be empty/quoted)
- `host`, `originalhost` (**extract**, but not when the criterion itself is negated)
- `user`, `group`, `localuser`, `localgroup`
- `localnetwork`
- `tagged`
- `sessiontype`
- `version`
- `all`, `canonical`, `final` (zero-arg / standalone forms)
- connection/context criteria such as `address`, `localaddress`, `ruser`, `rdomain`, `localport`, `remoteport` (consume arguments when known)

This list is not exhaustive; unknown criteria must be handled without aborting.

## Token alignment and argument arity

Match criteria vary in arity and form (single value, comma-separated values, boolean-like keywords, rest-of-line command strings). Tolerant token skipping must be good enough to recover `host`/`originalhost` without over-parsing.

Practical guidance:
- Use explicit criterion metadata when available: argument kind/arity, value/list forms, and rest-of-line command cases.
- Support `criterion=value`, `criterion = value`, and spaced forms. Preserve empty quoted argument tokens where relevant (e.g. `command ""`, `tagged ""`).
- When arity is unknown, prefer a conservative skip strategy that does not drop subsequent `host`/`originalhost` criteria if they can be identified; but never guess in a way that causes execution.
- `exec` command text must remain opaque/unexecuted; recovering additional host criteria after an `exec` argument on the same line may be unsafe. `command` should be aligned as a value argument (not arbitrary rest-of-line text) for parsing safety/alignment.

## Negation and lists

- Criteria and host values may be negated with `!` (e.g. `!host`, or `!name` patterns depending on form).
- Names may be comma-separated.
- Extraction should capture these forms as declared; completable filtering (dropping negated/wildcard-only patterns) is a separate step (see [host-candidates.md](host-candidates.md)).

## Examples (conceptual)

```sshconfig
# Must not execute; still index other declared hosts in file
Match exec "test -f /tmp/marker"
    Host internal-only

# Must extract host candidates from criteria
Match originalhost bastion,*.example.com
    HostName 10.0.0.1

Match host prod-*,!prod-db user alice
    User deploy

# Zero-arg / combined / other criteria must be tolerated
Match all
Match canonical all
Match localnetwork 10.0.0.0/8 user bob tagged deploy
```

(Examples illustrate extraction/tolerance intent; exact candidate filtering follows the consumer's documented rules.)

## Safety tests

Configs containing `Match exec "..."` must not create side effects. Tests should verify observable absence (e.g. temp marker file not created). See [testing.md](testing.md).

## Related

- [ssh-config-syntax.md](ssh-config-syntax.md) — directive recognition and line parsing.
- [host-candidates.md](host-candidates.md) — candidate filtering, dedupe, sorting.
- [option-metadata.md](option-metadata.md) — Match criterion metadata for better skipping.
- [overview.md](overview.md) — non-goals and guarantees.
