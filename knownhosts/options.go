package knownhosts

import (
	"os"
	"path/filepath"
)

// Option configures known_hosts indexing behavior.
type Option func(*indexOptions)

type indexOptions struct {
	userKnownHostsPath   string
	systemKnownHostsPath string
	homeDir              string
	readFile             func(string) ([]byte, error)
}

// WithUserKnownHostsPath overrides the default user known_hosts path (~/.ssh/known_hosts).
func WithUserKnownHostsPath(path string) Option {
	return func(o *indexOptions) { o.userKnownHostsPath = path }
}

// WithSystemKnownHostsPath overrides the default system known_hosts path (/etc/ssh/ssh_known_hosts).
func WithSystemKnownHostsPath(path string) Option {
	return func(o *indexOptions) { o.systemKnownHostsPath = path }
}

// WithHomeDir overrides the home directory used for default paths.
func WithHomeDir(path string) Option {
	return func(o *indexOptions) { o.homeDir = path }
}

// WithReadFile overrides file reading (primarily for tests).
func WithReadFile(fn func(string) ([]byte, error)) Option {
	return func(o *indexOptions) { o.readFile = fn }
}

func defaultOptions() indexOptions {
	home, _ := os.UserHomeDir()
	return indexOptions{
		userKnownHostsPath:   filepath.Join(home, ".ssh", "known_hosts"),
		systemKnownHostsPath: filepath.Join(string(filepath.Separator), "etc", "ssh", "ssh_known_hosts"),
		homeDir:              home,
		readFile:             os.ReadFile,
	}
}
