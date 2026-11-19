// Package knownhosts will provide parsing and host enrichment from OpenSSH
// known_hosts files for completion use cases.
//
// Planned capabilities (Phase 3): parse plain/hashed/bracketed host entries and
// markers, ignore unhasable hashed hosts for completion, and merge known hosts
// with ssh_config-declared candidates.
package knownhosts
