# carapace-ssh

`carapace-ssh` is a Go library for parsing and indexing OpenSSH-related data for shell completion. It provides a tolerant, completion-safe ssh_config indexer focused on host candidate extraction for `ssh`, `scp`, `sftp`, and related completers.

## Scope

- Parse and index ssh_config without evaluating runtime conditions.
- Extract literal host candidates from `Host` and `Match host|originalhost`.
- Follow `Include` directives with globbing, recursion, cycle detection, and `~` expansion.
- Tolerate unknown keywords/Match criteria and malformed lines where possible (best-effort recovery).
- Support explicit config paths (`-F` semantics) and default user/system config discovery.
- Expose best-effort static per-host facts without dynamic Match evaluation.
- Parse known_hosts files for completable host enrichment and merging (skips hashed and revoked hosts; extracts bracketed host/port forms to hostnames).
- Provide OpenSSH keyword and Match criterion metadata for tolerant parsing.
- Parse/format destination helpers for host/port, user@host, scp remote separators, ProxyJump hops, sftp/scp URI forms, and remote destination parsing.
- Provide completion-safe identity file discovery/pairing helpers.
- Provide tolerant `authorized_keys` parsing helpers.
- Provide an opt-in strict resolution path via `ssh -G` (outside default completion paths).

## Completion-safety guarantees

- **Never executes** `Match exec`, `Match command`, or any external process.
- No network access during parsing/indexing.
- Best-effort/tolerant behavior: unreadable includes or unknown constructs must not prevent returning results from the rest of the config graph; malformed/binary input is tolerated where practical.

This library is not an `ssh -G`-compatible evaluator and does not perform dynamic Match evaluation, canonicalization, or token expansion.

## Contract notes

- `HostCandidates()` returns literal completable hosts (deduped, sorted). `DeclaredHosts()` preserves declared tokens including patterns/negations before filtering.
- `isCompletableHost` rejects empty tokens, tokens starting with `!`, tokens containing whitespace, and tokens containing pattern/wildcard characters such as `* ? [ ] { } ( ) | + ^ $ !` (glob-like/negated forms are not offered as literal completions).
- In `Match`, only `host` and `originalhost` contribute extracted candidates; negated `host`/`originalhost` criteria do not contribute candidates. `exec` is treated as opaque/rest (never executed; extraction stops at it). `command` is treated as a value argument (consumed, never executed). Unknown criteria are skipped conservatively to preserve alignment where feasible.
- `Include` directives inside `Host`/`Match` blocks are followed conservatively for discovery (completion-safe over-inclusion).
- Static facts are best-effort and non-authoritative; in the current static model facts may be associated with declared negated `Host` tokens.
- `destination.ParseRemote()` tries URI parsing first and may fall back to tolerant SCP-style parsing; use `destination.ParseURI()` for URI-only acceptance.
- `option` metadata is curated for completion-safe parsing, not a full ssh_config specification.

## Usage

```go
package main

import (
	"fmt"

	"github.com/carapace-sh/carapace-ssh/config"
	"github.com/carapace-sh/carapace-ssh/knownhosts"
)

func main() {
	// Empty path uses default user/system configs; explicit path is isolated
	// (like -F) but follows its Include directives.
	hosts, _ := config.HostCandidates("")
	for _, host := range hosts {
		fmt.Println(host)
	}

	// Known hosts enrichment (best-effort; missing files are non-fatal).
	known, _ := knownhosts.IndexDefault()
	// Or index config-declared UserKnownHostsFile/GlobalKnownHostsFile paths:
	// cfg, _ := config.IndexConfig("")
	// known, _ = knownhosts.IndexFromConfigPaths(cfg.KnownHostsFiles())
	merged := knownhosts.Merge(hosts, known.HostCandidates())
	_ = merged
}
```

Explicit config path:

```go
index, err := config.IndexFile("/path/to/ssh_config")
hosts := index.HostCandidates()
declared := index.DeclaredHosts() // includes patterns/negations before filtering
facts := index.Facts("myhost")     // best-effort static directives
```

Other config helpers: `IndexConfig(path, ...)`, `IndexDefault(...)`, `IndexPaths(...)`, `IndexReader(...)`, `DefaultSources(...)`.

## Development

See [AGENTS.md](AGENTS.md) for development notes. An OpenSSH completion knowledge base is available in [skills/openssh/](skills/openssh/).

## License

MIT
