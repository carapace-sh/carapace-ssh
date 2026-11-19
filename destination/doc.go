// Package destination provides helpers for parsing and formatting SSH/SCP/SFTP
// destination forms used in completion contexts.
//
// Supported forms are handled on a best-effort basis and include user@host,
// host:port, bracketed IPv6 hosts, scp-style remote path separators, ProxyJump
// hop specifications, sftp/scp URI forms, and remote destination parsing.
package destination
