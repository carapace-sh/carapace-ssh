// Package resolve provides an opt-in strict host resolution path using OpenSSH
// client configuration evaluation.
//
// Unlike the default parsing/indexing APIs in this module, helpers in this
// package may execute an external ssh process (ssh -G) when explicitly invoked.
// They must not be used by default completion paths.
//
// Enabling strict resolution is more invasive than executing ssh alone: the
// evaluation performed by ssh -G also runs commands declared in the evaluated
// configuration, including Match exec directives, under the user's shell.
// Consumers should only use this package on configuration they trust and are
// encouraged to pass an explicit WithConfigPath and use HostContext with a
// deadline.
package resolve
