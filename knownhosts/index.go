package knownhosts

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
)

// Index contains parsed completable host candidates from known_hosts sources.
type Index struct {
	hosts map[string]struct{}
}

func newIndex() *Index {
	return &Index{hosts: map[string]struct{}{}}
}

// HostCandidates returns filtered, deduplicated, sorted host candidates.
func (i *Index) HostCandidates() []string {
	if i == nil {
		return []string{}
	}
	return sortedUnique(i.hosts)
}

// IndexFile indexes a known_hosts file. Tilde-prefixed paths are expanded
// using configured home directory. Missing or unreadable files are treated as
// non-fatal.
func IndexFile(path string, opts ...Option) (*Index, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	idx := newIndex()
	if path == "" {
		return idx, nil
	}
	path = expandTilde(path, o.homeDir)
	content, err := o.readFile(path)
	if err != nil {
		if isNonFatalReadError(err) {
			return idx, nil
		}
		return idx, err
	}
	_ = parseReader(bytes.NewReader(content), func(host string) {
		idx.hosts[host] = struct{}{}
	})
	return idx, nil
}

// IndexPaths indexes multiple known_hosts files in order. Tilde-prefixed paths
// are expanded using configured home directory. Missing or unreadable files are
// treated as non-fatal.
func IndexPaths(paths []string, opts ...Option) (*Index, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	idx := newIndex()
	var errs []error
	for _, path := range paths {
		if path == "" {
			continue
		}
		path = expandTilde(path, o.homeDir)
		content, err := o.readFile(path)
		if err != nil {
			if isNonFatalReadError(err) {
				continue
			}
			errs = append(errs, err)
			continue
		}
		_ = parseReader(bytes.NewReader(content), func(host string) {
			idx.hosts[host] = struct{}{}
		})
	}
	if len(errs) > 0 {
		return idx, errors.Join(errs...)
	}
	return idx, nil
}

func isNonFatalReadError(err error) bool {
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission)
}

// IndexDefault indexes default user and system known_hosts files.
func IndexDefault(opts ...Option) (*Index, error) {
	return IndexPaths(DefaultPaths(opts...), opts...)
}

// IndexFromConfigPaths indexes known_hosts files from config-declared path
// values. Empty and case-insensitive "none" entries are ignored.
func IndexFromConfigPaths(paths []string, opts ...Option) (*Index, error) {
	filtered := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == "" || strings.EqualFold(p, "none") {
			continue
		}
		filtered = append(filtered, p)
	}
	return IndexPaths(filtered, opts...)
}

// IndexReader indexes known_hosts content from a reader.
func IndexReader(r io.Reader) (*Index, error) {
	idx := newIndex()
	if r == nil {
		return idx, nil
	}
	if err := parseReader(r, func(host string) {
		idx.hosts[host] = struct{}{}
	}); err != nil {
		return idx, err
	}
	return idx, nil
}
