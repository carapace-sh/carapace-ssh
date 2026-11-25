// Package config provides a tolerant, completion-safe ssh_config indexer.
//
// It extracts completable host candidates from Host directives and Match
// host/originalhost criteria, follows Include directives, and supports explicit
// config paths as well as default user/system config discovery.
//
// HostCandidates returns literal completable hosts (no patterns/wildcards,
// negations, or whitespace). DeclaredHosts preserves declared Host/Match host
// tokens as written/normalized for inspection, including patterns and negations.
// Static facts are associated on a best-effort basis and are not ssh -G
// equivalent.
//
// Parsing and indexing do not evaluate runtime conditions and never execute
// Match exec or Match command directives. Unknown keywords/criteria, malformed
// lines, and unreadable includes are handled on a best-effort basis so results
// from the rest of the config graph can still be returned.
//
// This package is intended for shell completion use cases and does not aim for
// ssh -G parity, canonicalization, or dynamic token expansion.
package config
