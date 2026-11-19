# carapace-ssh

`carapace-ssh` is a Go library for parsing and indexing OpenSSH-related data for shell completion. It provides a tolerant, completion-safe ssh_config indexer focused on host candidate extraction for `ssh`, `scp`, `sftp`, and related completers.

## Scope (MVP)

- Parse and index ssh_config without evaluating runtime conditions.
- Extract literal host candidates from `Host` and `Match host|originalhost`.
- Follow `Include` directives with globbing, recursion, cycle detection, and `~` expansion.
- Tolerate unknown keywords/Match criteria and malformed lines where possible.
- Support explicit config paths (`-F` semantics) and default user/system config discovery.

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
)

func main() {
	index, err := config.IndexDefault()
	if err != nil {
		// best-effort usage may still inspect partial results depending on API evolution
		fmt.Println(err)
		return
	}

	for _, host := range index.HostCandidates() {
		fmt.Println(host)
	}
}
```

Explicit config path (isolated, follows its includes only):

```go
index, err := config.IndexFile("/path/to/ssh_config")
```

## Development

See [plan.md](plan.md) for the implementation roadmap and [AGENTS.md](AGENTS.md) for development notes. An OpenSSH completion knowledge base is available in [skills/openssh/](skills/openssh/).

## License

MIT
