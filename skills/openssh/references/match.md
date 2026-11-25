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

Include comma-separated lists. Do **not** extract hosts from negated `host`/`originalhost` criteria (e.g. `!host bad.example`); their arguments/values are still consumed for alignment when parsing the criterion form. Extraction is static/best-effort and must not rely on runtime evaluation.

## Behavioral requirements (tolerant, non-executing)

| Requirement | Detail |
|-------------|--------|
| **No execution** | Never run `exec`/`command` or any external process. Treat `exec` as opaque/rest-of-criterion text and stop host extraction for that `Match` line to avoid misparsing; `command` takes a value argument (possibly empty/quoted) and must also never be executed (consume it for alignment). |
| **Zero-arg criteria** | Parse/skip known zero-arg criteria such as `all`, `canonical`, `final` (and `canonical all` forms as encountered). |
| **Argument-consuming criteria** | Parse/skip criteria that consume arguments so token alignment is preserved well enough to still find `host`/`originalhost` in the same `Match` line when safe. |
| **Unknown criteria/keywords** | Must not abort parsing of the file or prevent host extraction elsewhere. Fall back safely (conservative skipping/alignment). |
| **Host extraction survives gating** | Hosts declared in blocks gated by non-evaluable/unsupported criteria must still be discoverable as declared candidates where applicable to extraction goals (especially `Match host|originalhost` values themselves). |

## Known criteria to account for

Tolerant handling commonly needs to account for criteria including:
- `exec` (command text; **never execute**, opaque; stop extraction on this criterion for the line)
- `command` (requested command/subsystem pattern argument; **never execute**; value arg, may be empty/quoted)
- `host`, `originalhost` (**extract when not negated**, e.g. `!host ...` is skipped)
- `user`, `group`, `localuser`, `localgroup`
- `localnetwork`
- `tagged`
- `sessiontype`
- `version`
- `all`, `canonical`, `final` (zero-arg / standalone forms)
- connection/context criteria such as `address`, `localaddress`, `ruser`, `rdomain`, `localport`, `remoteport` (consume arguments when known)

For completion-safe parsing, treat documented value-arg criteria as consuming a single argument (including `criterion=value`, spaced `=`, and empty quoted forms) unless a criterion is explicitly command-like (`exec` rest/opaque stop behavior; `command` value arg, never executed).

This list is not exhaustive; unknown criteria must be handled without aborting.

## Token alignment and argument arity

Match criteria vary in arity and form. Tolerant skipping should recover `host`/`originalhost` where practical without executing anything.

Practical guidance:
- Use explicit criterion metadata when available (`MatchArgNone`, `MatchArgValue`, `MatchArgRest`).
- Support `criterion=value`, `criterion = value`, and spaced forms. Preserve empty quoted argument tokens where relevant (e.g. `command ""`, `tagged ""`).
- When arity is unknown, prefer conservative skipping that avoids dropping later identifiable `host`/`originalhost` criteria when possible; never guess in a way that causes execution.
- `exec` stops further host extraction for the same `Match` line in this library's current behavior (opaque). `command` is aligned as a value argument.

## Negation and lists

- Criteria may be negated with `!` (e.g. `!host ...`). Do not extract host candidates from negated `host`/`originalhost` criteria.
- Names may be comma-separated.
- Extracted forms are then subject to completable filtering separately (see [host-candidates.md](host-candidates.md)).

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
