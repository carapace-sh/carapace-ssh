// Package knownhosts provides parsing and host enrichment from OpenSSH
// known_hosts files for completion use cases.
//
// It parses plain, comma-separated, and bracketed [host]:port entries (including
// bracketed IPv6 forms, returned as raw hostnames), tolerates @cert-authority and
// @revoked markers, ignores hashed hosts for completion, and excludes @revoked
// entries from completable candidates. Explicit literal @cert-authority host
// tokens may be included when they are otherwise completable; wildcard/negated
// tokens are not. It supports merging known hosts with ssh_config-declared
// candidates. Tilde-prefixed input paths are expanded using configured home
// directory. Missing or unreadable files are treated as non-fatal by indexing
// helpers.
package knownhosts
