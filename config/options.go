package config

import (
	"os"
	"path/filepath"
)

// Source describes a config entry point and its resolution context.
type Source struct {
	Path    string
	BaseDir string
	System  bool
}

// Option configures indexing behavior.
type Option func(*indexOptions)

type indexOptions struct {
	userConfigPath   string
	userConfigSet    bool
	systemConfigPath string
	homeDir          string
	readFile         func(string) ([]byte, error)
	glob             func(string) ([]string, error)
	abs              func(string) (string, error)
	evalSymlinks     func(string) (string, error)
}

// WithUserConfigPath overrides the default user config path (~/.ssh/config).
func WithUserConfigPath(path string) Option {
	return func(o *indexOptions) {
		o.userConfigPath = path
		o.userConfigSet = true
	}
}

// WithSystemConfigPath overrides the default system config path (/etc/ssh/ssh_config).
func WithSystemConfigPath(path string) Option {
	return func(o *indexOptions) { o.systemConfigPath = path }
}

// WithHomeDir overrides the home directory used for ~ expansion and default paths.
func WithHomeDir(path string) Option {
	return func(o *indexOptions) { o.homeDir = path }
}

// WithReadFile overrides file reading (primarily for tests).
func WithReadFile(fn func(string) ([]byte, error)) Option {
	return func(o *indexOptions) { o.readFile = fn }
}

// WithGlob overrides filepath.Glob (primarily for tests).
func WithGlob(fn func(string) ([]string, error)) Option {
	return func(o *indexOptions) { o.glob = fn }
}

// WithAbs overrides filepath.Abs (primarily for tests).
func WithAbs(fn func(string) (string, error)) Option {
	return func(o *indexOptions) { o.abs = fn }
}

// WithEvalSymlinks overrides filepath.EvalSymlinks (primarily for tests).
func WithEvalSymlinks(fn func(string) (string, error)) Option {
	return func(o *indexOptions) { o.evalSymlinks = fn }
}

func defaultOptions() indexOptions {
	home, _ := os.UserHomeDir()
	return indexOptions{
		systemConfigPath: filepath.Join(string(filepath.Separator), "etc", "ssh", "ssh_config"),
		homeDir:          home,
		readFile:         os.ReadFile,
		glob:             filepath.Glob,
		abs:              filepath.Abs,
		evalSymlinks:     filepath.EvalSymlinks,
	}
}

func (o indexOptions) defaultUserSource() Source {
	base := filepath.Join(o.homeDir, ".ssh")
	path := o.userConfigPath
	if !o.userConfigSet {
		path = filepath.Join(base, "config")
	}
	if base == ".ssh" && o.homeDir == "" {
		if path != "" {
			base = filepath.Dir(path)
		}
	}
	return Source{Path: path, BaseDir: base, System: false}
}

func (o indexOptions) defaultSystemSource() Source {
	return Source{Path: o.systemConfigPath, BaseDir: filepath.Dir(o.systemConfigPath), System: true}
}

// DefaultSources returns default user and system config sources in order.
func DefaultSources(opts ...Option) []Source {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	sources := make([]Source, 0, 2)
	if s := o.defaultUserSource(); s.Path != "" {
		sources = append(sources, s)
	}
	if s := o.defaultSystemSource(); s.Path != "" {
		sources = append(sources, s)
	}
	return sources
}
