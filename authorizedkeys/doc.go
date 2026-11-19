// Package authorizedkeys provides tolerant parsing helpers for OpenSSH
// authorized_keys files.
//
// Parsing is completion-safe: helpers read and parse data only, without executing
// external programs or accessing the network.
package authorizedkeys
