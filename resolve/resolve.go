package resolve

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Result contains resolved ssh configuration values for a host.
type Result struct {
	Host   string
	values map[string][]string
	order  []string
}

// First returns the first value for a canonicalized key, or empty string.
func (r *Result) First(key string) string {
	if r == nil {
		return ""
	}
	values := r.values[canonicalKey(key)]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// All returns all values for a canonicalized key.
func (r *Result) All(key string) []string {
	if r == nil {
		return nil
	}
	values := r.values[canonicalKey(key)]
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

// Keys returns canonicalized keys present in the result in deterministic order
// of first appearance.
func (r *Result) Keys() []string {
	if r == nil {
		return nil
	}
	keys := make([]string, len(r.order))
	copy(keys, r.order)
	return keys
}

type resultBuilder struct {
	host   string
	values map[string][]string
	order  []string
	seen   map[string]struct{}
}

func newResultBuilder(host string) *resultBuilder {
	return &resultBuilder{
		host:   host,
		values: map[string][]string{},
		seen:   map[string]struct{}{},
	}
}

func (b *resultBuilder) add(key, value string) {
	key = canonicalKey(key)
	if key == "" {
		return
	}
	if _, ok := b.seen[key]; !ok {
		b.seen[key] = struct{}{}
		b.order = append(b.order, key)
	}
	b.values[key] = append(b.values[key], value)
}

func (b *resultBuilder) done() *Result {
	return &Result{Host: b.host, values: b.values, order: b.order}
}

// Runner executes ssh -G and returns its standard output.
type Runner func(ctx context.Context, sshPath string, args []string) ([]byte, error)

type options struct {
	sshPath    string
	configPath string
	runner     Runner
}

// Option configures strict host resolution.
type Option func(*options)

// WithSSHPath sets the ssh executable path or name (default: "ssh").
func WithSSHPath(path string) Option {
	return func(o *options) { o.sshPath = path }
}

// WithConfigPath sets an explicit ssh_config path passed as -F.
func WithConfigPath(path string) Option {
	return func(o *options) { o.configPath = path }
}

// WithRunner overrides command execution (primarily for tests).
func WithRunner(runner Runner) Option {
	return func(o *options) { o.runner = runner }
}

// Host resolves a host using `ssh -G`. This is opt-in and executes an external
// process; it must not be used by default completion paths.
func Host(host string, opts ...Option) (*Result, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("resolve: host is required")
	}

	o := options{sshPath: "ssh", runner: defaultRunner}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if strings.TrimSpace(o.sshPath) == "" {
		o.sshPath = "ssh"
	}
	if o.runner == nil {
		o.runner = defaultRunner
	}

	args := make([]string, 0, 5)
	if strings.TrimSpace(o.configPath) != "" {
		args = append(args, "-F", o.configPath)
	}
	args = append(args, "-G", "--", host)

	out, err := o.runner(context.Background(), o.sshPath, args)
	if err != nil {
		return nil, fmt.Errorf("resolve: ssh -G: %w", err)
	}
	return parseConfig(out, host), nil
}

func defaultRunner(ctx context.Context, sshPath string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, sshPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return out, nil
}

func parseConfig(out []byte, host string) *Result {
	b := newResultBuilder(host)
	for _, line := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		key, value, ok := splitConfigLine(line)
		if !ok {
			continue
		}
		b.add(key, value)
	}
	return b.done()
}

func splitConfigLine(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}
	idx := strings.IndexAny(line, " \t")
	if idx < 0 {
		return strings.ToLower(line), "", true
	}
	key = strings.ToLower(strings.TrimSpace(line[:idx]))
	value = strings.TrimSpace(line[idx+1:])
	return key, value, true
}

func canonicalKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}
