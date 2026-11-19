// Package key provides completion-safe helpers for discovering SSH identity
// files, pairing related key material, and locating agent socket hints.
//
// Helpers inspect filesystem paths and environment values only; they do not
// execute external programs, probe sockets, access the network, or read private
// key contents.
package key
