// Package knownhosts provides parsing and host enrichment from OpenSSH
// known_hosts files for completion use cases.
//
// It parses plain, comma-separated, and bracketed [host]:port entries, tolerates
// @cert-authority/@revoked markers, ignores hashed hosts for completion, and
// supports merging known hosts with ssh_config-declared candidates. Missing or
// unreadable files are treated as non-fatal by indexing helpers.
package knownhosts
