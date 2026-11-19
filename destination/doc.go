// Package destination provides helpers for parsing and formatting SSH/SCP
// destination forms used in completion contexts.
//
// Supported forms are handled on a best-effort basis and include user@host,
// host:port, bracketed IPv6 hosts, and scp-style remote path separators.
package destination
