# carapace-ssh

`carapace-ssh` is a Go library for parsing and indexing OpenSSH-related data for shell completion. It provides a tolerant, completion-safe ssh_config indexer focused on host candidate extraction for `ssh`, `scp`, `sftp`, and related completers.

## Scope

- Parse and index ssh_config without evaluating runtime conditions.
- Extract literal host candidates from `Host` and `Match host|originalhost`.
- Follow `Include` directives with globbing, recursion, cycle detection, and `~` expansion.
- Tolerate unknown keywords/Match criteria and malformed lines where possible.
- Support explicit config paths (`-F` semantics) and default user/system config discovery.
- Parse known_hosts files for completable host enrichment (initial Phase 3 support).

## Completion-safety guarantees

- **Never executes** `Match exec`, `Match command`, or any external process.
- No network access during parsing/indexing.
- Best-effort/tolerant behavior: unreadable includes or unknown constructs must not prevent returning results from the rest of the config graph.

This library is not an `ssh -G`-compatible evaluator and does not perform dynamic Match evaluation, canonicalization, or token expansion.

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
	merged := knownhosts.Merge(hosts, known.HostCandidates())
	_ = merged
}
```

Explicit config path:

```go
index, err := config.IndexFile("/path/to/ssh_config")
hosts := index.HostCandidates()
declared := index.DeclaredHosts() // includes patterns/negations before filtering
```

Other config helpers: `IndexConfig(path, ...)`, `IndexDefault(...)`, `IndexPaths(...)`, `IndexReader(...)`.

## Development

See [plan.md](plan.md) for the implementation roadmap and [AGENTS.md](AGENTS.md) for development notes. An OpenSSH completion knowledge base is available in [skills/openssh/](skills/openssh/).

## License

MIT
