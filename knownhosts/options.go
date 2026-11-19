package knownhosts

import (
	"os"
	"path/filepath"
)

// Option configures known_hosts indexing behavior.
type Option func(*indexOptions)

type indexOptions struct {
	userKnownHostsPath   string
	userKnownHostsSet    bool
	systemKnownHostsPath string
	homeDir              string
	readFile             func(string) ([]byte, error)
}

// WithUserKnownHostsPath overrides the default user known_hosts path (~/.ssh/known_hosts).
func WithUserKnownHostsPath(path string) Option {
	return func(o *indexOptions) {
		o.userKnownHostsPath = path
		o.userKnownHostsSet = true
	}
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

// DefaultPaths returns default user and system known_hosts paths in order.
func DefaultPaths(opts ...Option) []string {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	paths := make([]string, 0, 2)
	if p := o.defaultUserKnownHostsPath(); p != "" {
		paths = append(paths, p)
	}
	if o.systemKnownHostsPath != "" {
		paths = append(paths, o.systemKnownHostsPath)
	}
	return paths
}

func defaultOptions() indexOptions {
	home, _ := os.UserHomeDir()
	return indexOptions{
		systemKnownHostsPath: filepath.Join(string(filepath.Separator), "etc", "ssh", "ssh_known_hosts"),
		homeDir:              home,
		readFile:             os.ReadFile,
	}
}

func (o indexOptions) defaultUserKnownHostsPath() string {
	if o.userKnownHostsSet {
		return o.userKnownHostsPath
	}
	base := o.homeDir
	if base == "" {
		return ""
	}
	return filepath.Join(base, ".ssh", "known_hosts")
}
