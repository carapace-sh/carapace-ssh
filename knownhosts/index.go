package knownhosts

import (
	"bytes"
	"errors"
	"io"
	"os"
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

// IndexFile indexes a known_hosts file. Missing files are treated as non-fatal.
func IndexFile(path string, opts ...Option) (*Index, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	idx := newIndex()
	if path == "" {
		return idx, nil
	}
	content, err := o.readFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return idx, nil
		}
		return idx, err
	}
	_ = parseReader(bytes.NewReader(content), func(host string) {
		idx.hosts[host] = struct{}{}
	})
	return idx, nil
}

// IndexPaths indexes multiple known_hosts files in order. Missing files are
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
		content, err := o.readFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
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

// IndexDefault indexes default user and system known_hosts files.
func IndexDefault(opts ...Option) (*Index, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return IndexPaths([]string{o.userKnownHostsPath, o.systemKnownHostsPath},
		WithHomeDir(o.homeDir),
		WithReadFile(o.readFile),
		WithUserKnownHostsPath(o.userKnownHostsPath),
		WithSystemKnownHostsPath(o.systemKnownHostsPath),
	)
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
