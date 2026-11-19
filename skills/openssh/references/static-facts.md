# Static Per-Host Facts

Extraction of useful declared directives per host/block without dynamic evaluation (Phase 4). Enables richer completion (identity files, users, ports, jump hosts) while staying completion-safe.

> **Source of truth**: [plan.md](../../../plan.md), Phase 4.

## Goal

The ssh_config index/model retains statically present directive facts and exposes query APIs (`Facts`, `HostsWithFacts`, `KnownHostsFiles`) for best-effort per-host information. This does **not** attempt `ssh -G` parity or runtime condition evaluation.

## Candidate directives to retain (initial)

| Directive | Completion relevance |
|-----------|----------------------|
| `HostName` | Effective destination target hints; may influence completion context. |
| `User` | Suggest `user@` prefixes / defaults. |
| `Port` | Suggest default/nonstandard ports. |
| `IdentityFile` | Host-aware identity suggestions for `-i` / `IdentityFile`. |
| `CertificateFile` | Host-aware certificate suggestions. |
| `ProxyJump` | Suggest jump hosts / multi-hop targets. |
| `ProxyCommand` | Context for proxy-related completion (static presence/value). |
| `IdentityAgent` | Declared agent socket/path hint for agent-aware completion context. |
| `UserKnownHostsFile` | Locate known_hosts sources per host context (when statically declared). |
| `GlobalKnownHostsFile` | Locate system known_hosts sources per host context. |

Additional directives may be added as consumers need them, but keep scope tied to completion value.

## Association semantics (best-effort/static)

Defining how directives associate with hosts is critical and inherently approximate without evaluation:

- Facts may be associated with **declared blocks** (`Host`/`Match` sections) and/or exposed via **simple matching helpers**.
- Multiple blocks/patterns may match a given name in OpenSSH with first-match-wins and conditional semantics; static extraction cannot fully reproduce effective resolution.
- `Match`-gated blocks complicate association: the indexer does not evaluate whether a Match applies, but it can still record what directives are declared under it.

**API/behavior must be explicit about limitations vs `ssh -G`.** Consumers should treat results as hints/suggestions derived from declarations, not authoritative resolved configuration.

## No dynamic evaluation

Do not:
- expand tokens (`%h`, `%r`, `%p`, `%u`, etc.) at runtime connection context
- evaluate `Match exec`/`command` or dynamic criteria to decide applicability
- perform canonicalization
- execute `ssh -G` by default

An opt-in strict resolution path using `ssh -G` may be considered later (Phase 6+) but must never be the default completion path.

## Consumer use cases

- **Identity/certificate suggestions**: when completing `-i` or values related to identity, prefer files declared for the target host when known.
- **Destination hints**: suggest `user@host` forms, default ports, or HostName-aware targets where helpful.
- **Jump/proxy sources**: better `-J` / `ProxyJump` target suggestions from declared config.
- **known_hosts discovery**: use declared `UserKnownHostsFile`/`GlobalKnownHostsFile` to enrich host sources.

## Data model considerations

- Preserve enough structure to answer "what facts are declared in blocks that mention/target X?" without claiming full resolution.
- Keep results deterministic (stable ordering where collections are returned).
- Consider representing values as declared strings plus parsed forms where cheap (e.g. port numbers), but avoid over-normalizing in ways that lose meaning.

## Testing expectations

- Extraction of listed directives from `Host` and `Match` blocks.
- Multiple declarations, ordering as declared (if exposed), and deterministic query results.
- Coexistence with tolerant Match parsing (facts still collected when blocks contain non-evaluable criteria, subject to defined association rules).
- Clear tests/documentation for cases where static association is ambiguous.

## Related

- [match.md](match.md) — Match blocks are parsed without evaluation.
- [host-candidates.md](host-candidates.md) — host identity/candidate extraction.
- [known-hosts.md](known-hosts.md) — per-host known_hosts file directives feed enrichment.
- [integration.md](integration.md) — where carapace-bin may adopt richer actions.
