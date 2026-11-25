// Package option provides OpenSSH config keyword and Match criterion metadata
// for parsing and completion.
//
// It exposes keyword and Match criterion lookup helpers, including argument kind
// and host-extraction metadata used by tolerant ssh_config parsing. Metadata is
// intentionally curated for completion-safe parsing: keyword coverage is a small
// subset, while Match criteria focus on safe alignment and host extraction.
package option
