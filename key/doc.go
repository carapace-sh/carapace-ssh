// Package key provides completion-safe helpers for discovering SSH identity
// files and pairing related key material.
//
// Helpers inspect filesystem paths only; they do not execute external programs,
// access the network, or read private key contents.
package key
