# Host Candidates

How to collect, filter, deduplicate, and expose host candidates from ssh_config for completion.

## Sources of declared candidates

Collect literal host names/pattern tokens from:

1. **`Host` arguments**
   - Each token in a `Host` line is a declared host pattern/name.
2. **`Match host|originalhost` criteria**
   - Values associated with `host` and `originalhost` criteria in `Match` lines.
   - May be comma-separated lists.
   - May include `!`-negated forms as they appear in criteria arguments.

Extraction must occur even when the surrounding block is gated by non-evaluable/unsupported Match criteria (e.g. `Match exec`, `Match user`), as long as the `host`/`originalhost` values themselves are statically present and safely parseable. See [match.md](match.md).

## Raw vs filtered candidates

Expose raw and/or filtered candidates as the API evolves. At minimum, provide a filtered, completable set.

### Filtered (completable) candidates

The filtered set must exclude non-completable patterns/wildcards and negations. Common guidance:
- exclude tokens containing glob/wildcard/regex-like or pattern characters or whitespace as appropriate (e.g. characters in the set `* ? [ ] { } ( ) | + ^ $ !` and whitespace, consistent with current behavior)
- correctly handle `!name` negations (negated entries are not completable hostnames)
- do not produce empty/invalid candidates

Current `carapace-ssh/config` contract: `HostCandidates()` returns this literal completable set (deduped, sorted). `DeclaredHosts()` preserves declared Host/Match host tokens including patterns/negations before filtering. Filtering rules should be documented and tested explicitly so consumers can rely on stable behavior.

### Raw/declared candidates

Richer APIs (e.g. `DeclaredHosts()`) may expose declared forms, block/source context, or pattern tokens for future consumers (static facts, diagnostics). Keep public surface small initially; add only when needed.

## Determinism

Public results must be:
- **deduplicated** (a host discovered via multiple files/blocks appears once in the filtered set)
- **stably ordered** where appropriate (e.g. sorted) to avoid order dependence on map iteration

## API sketch

```go
func (idx *Index) HostCandidates() []string // filtered, deduped, sorted
// and/or DeclaredHosts() for richer info
```

Index construction APIs (illustrative): `IndexFile`, `IndexPaths`, `IndexDefault` with options for context/resolvers. See [defaults.md](defaults.md).

## Interaction with config graph

Candidates are aggregated across the indexed config graph:
- entry file(s) per source mode (explicit `-F` or defaults)
- all followed `Include` targets (respecting cycles and non-fatal errors)

See [include.md](include.md) and [defaults.md](defaults.md).

## Known limitations (by design)

- No dynamic evaluation: candidates come from declared syntax, not from `ssh -G`-resolved effective configuration.
- Patterns/wildcards are generally not offered as literal completions (filtered out), unless explicitly decided otherwise.
- Token expansion (`%h`, `%r`, `%p`) is out of scope; do not synthesize expanded candidates.
- IPv6/bracketed forms: raw IPv6 literals may pass the current literal filter and be completable when declared as plain tokens; bracketed `[host]:port` forms are not literal completion candidates. `knownhosts` extraction normalizes bracketed host:port entries to raw hostnames for completion, and `destination` helpers should be used for bracketed/port-qualified formatting.

## Testing focus

- `Host` argument extraction (single/multiple, mixed case, `=` forms).
- `Match host` / `Match originalhost` extraction, including comma-separated names and negated forms.
- Presence of candidates when config contains `Match exec`, `Match command`, or other non-evaluable/unknown criteria.
- `Match all` / `canonical all` tolerance.
- Filtering of wildcards/pattern chars and negations; dedupe/sort stability.

See [testing.md](testing.md).

## Related

- [match.md](match.md) — safe extraction from Match criteria.
- [ssh-config-syntax.md](ssh-config-syntax.md) — parsing/tolerance.
- [integration.md](integration.md) — host suggestion contracts for completers.
- [known-hosts.md](known-hosts.md) — additional host source to merge later.
