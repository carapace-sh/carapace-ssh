// Package resolve provides an opt-in strict host resolution path using OpenSSH
// client configuration evaluation.
//
// Unlike the default parsing/indexing APIs in this module, helpers in this
// package may execute an external ssh process (ssh -G) when explicitly invoked.
// They must not be used by default completion paths.
package resolve
