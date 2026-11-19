package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Index contains parsed host candidates from ssh_config sources.
type Index struct {
	hosts map[string]struct{}
}

// HostCandidates returns filtered, deduplicated, sorted host candidates.
func (i *Index) HostCandidates() []string {
	if i == nil {
		return []string{}
	}
	return sortedUniqueHosts(i.hosts)
}

// IndexFile indexes a single explicit config file as the entry point.
// It does not implicitly merge user/system defaults, but follows Include directives
// found in the file.
func IndexFile(path string, opts ...Option) (*Index, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	idx := &Index{hosts: map[string]struct{}{}}
	if path == "" {
		return idx, nil
	}

	source := makeSource(path, "", o)
	err := indexSource(idx, source, o)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return idx, nil
	}
	return idx, err
}

// IndexDefault indexes default user and system ssh_config entry points.
func IndexDefault(opts ...Option) (*Index, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	idx := &Index{hosts: map[string]struct{}{}}
	var errs []error
	for _, source := range []Source{o.defaultUserSource(), o.defaultSystemSource()} {
		if err := indexSource(idx, source, o); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return idx, errors.Join(errs...)
	}
	return idx, nil
}

// IndexPaths indexes multiple sources in order.
func IndexPaths(sources []Source, opts ...Option) (*Index, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	idx := &Index{hosts: map[string]struct{}{}}
	var errs []error
	for _, source := range sources {
		if source.Path != "" {
			source = makeSource(source.Path, source.BaseDir, o)
		} else if source.BaseDir == "" && !source.System && o.homeDir != "" {
			source.BaseDir = filepath.Join(o.homeDir, ".ssh")
		}
		if err := indexSource(idx, source, o); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return idx, errors.Join(errs...)
	}
	return idx, nil
}

// IndexReader indexes ssh_config content from a reader using the provided source
// context for include resolution. The source Path, when set, is used to avoid
// re-indexing the same logical file in include cycles.
func IndexReader(r io.Reader, source Source, opts ...Option) (*Index, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	idx := &Index{hosts: map[string]struct{}{}}
	if r == nil {
		return idx, nil
	}
	content, err := io.ReadAll(r)
	if err != nil {
		return idx, err
	}
	ix := &indexer{opts: o, visited: map[string]struct{}{}}

	if source.Path != "" {
		source = makeSource(source.Path, source.BaseDir, o)
		if resolved, resolveErr := ix.resolvePath(source.Path, filepath.Dir(source.Path), source); resolveErr == nil {
			ix.visited[resolved.visitKey] = struct{}{}
			source.Path = resolved.readPath
			if source.BaseDir == "" {
				source.BaseDir = filepath.Dir(resolved.readPath)
			}
		} else if source.BaseDir == "" {
			source.BaseDir = filepath.Dir(source.Path)
		}
	} else if source.BaseDir == "" {
		if !source.System && o.homeDir != "" {
			source.BaseDir = filepath.Join(o.homeDir, ".ssh")
		} else if source.System {
			source.BaseDir = filepath.Dir(o.systemConfigPath)
		}
	}

	ix.indexContent(idx, content, source)
	return idx, nil
}

func makeSource(path, baseDir string, o indexOptions) Source {
	expanded := expandTilde(path, o.homeDir)
	absPath, err := o.abs(expanded)
	if err != nil {
		absPath = filepath.Clean(expanded)
	}
	if baseDir == "" {
		baseDir = filepath.Dir(absPath)
	}
	return Source{Path: absPath, BaseDir: baseDir}
}

type indexer struct {
	opts    indexOptions
	visited map[string]struct{}
}

func indexSource(idx *Index, source Source, opts indexOptions) error {
	ix := &indexer{opts: opts, visited: map[string]struct{}{}}
	return ix.indexFile(idx, source.Path, source, false)
}

func (ix *indexer) indexFile(idx *Index, path string, ctx Source, missingOK bool) error {
	if path == "" {
		return nil
	}
	resolved, err := ix.resolvePath(path, ctx.BaseDir, ctx)
	if err != nil {
		if missingOK {
			return nil
		}
		return err
	}

	visitKey := resolved.visitKey
	if _, ok := ix.visited[visitKey]; ok {
		return nil
	}
	ix.visited[visitKey] = struct{}{}

	content, err := ix.opts.readFile(resolved.readPath)
	if err != nil {
		if missingOK || errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	ix.indexContent(idx, content, Source{
		Path:    resolved.readPath,
		BaseDir: filepath.Dir(resolved.readPath),
		System:  ctx.System,
	})
	return nil
}

func (ix *indexer) indexContent(idx *Index, content []byte, ctx Source) {
	directives := parseDirectives(content)
	for _, d := range directives {
		switch d.keyword {
		case "host":
			for _, arg := range d.args {
				for _, token := range splitCommaList(arg) {
					token = normalizeHostToken(token)
					if isCompletableHost(token) {
						idx.hosts[token] = struct{}{}
					}
				}
			}
		case "match":
			for _, token := range extractMatchHosts(d.args) {
				token = normalizeHostToken(token)
				if isCompletableHost(token) {
					idx.hosts[token] = struct{}{}
				}
			}
		case "include":
			ix.handleInclude(idx, d.args, Source{
				Path:    ctx.Path,
				BaseDir: ctx.BaseDir,
				System:  ctx.System,
			})
		}
	}
}

type resolvedPath struct {
	readPath string
	visitKey string
}

func (ix *indexer) resolvePath(path string, baseDir string, ctx Source) (resolvedPath, error) {
	expanded := expandTilde(path, ix.opts.homeDir)
	if !filepath.IsAbs(expanded) {
		if baseDir != "" {
			expanded = filepath.Join(baseDir, expanded)
		} else if !ctx.System && ix.opts.homeDir != "" {
			expanded = filepath.Join(ix.opts.homeDir, ".ssh", expanded)
		}
	}
	absPath, err := ix.opts.abs(expanded)
	if err != nil {
		absPath = filepath.Clean(expanded)
	}
	visitKey := absPath
	if real, err := ix.opts.evalSymlinks(absPath); err == nil {
		visitKey = real
	}
	return resolvedPath{readPath: absPath, visitKey: visitKey}, nil
}

func (ix *indexer) handleInclude(idx *Index, args []string, includeCtx Source) {
	for _, arg := range args {
		pattern := arg
		if pattern == "" {
			continue
		}
		expandedPattern := expandTilde(pattern, ix.opts.homeDir)
		if !filepath.IsAbs(expandedPattern) {
			if includeCtx.BaseDir != "" {
				expandedPattern = filepath.Join(includeCtx.BaseDir, expandedPattern)
			} else if !includeCtx.System && ix.opts.homeDir != "" {
				expandedPattern = filepath.Join(ix.opts.homeDir, ".ssh", expandedPattern)
			}
		}

		matches, err := ix.opts.glob(expandedPattern)
		if err != nil || len(matches) == 0 {
			// Missing/unreadable/non-matching includes are non-fatal.
			if !hasGlobMeta(expandedPattern) {
				_ = ix.indexFile(idx, expandedPattern, includeCtx, true)
			}
			continue
		}
		sort.Strings(matches)
		for _, match := range matches {
			_ = ix.indexFile(idx, match, includeCtx, true)
		}
	}
}

func expandTilde(path string, home string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	if home == "" {
		return path
	}
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		return filepath.Join(home, path[2:])
	}
	// ~user/... forms are not expanded without user lookup; return as-is.
	return path
}

func hasGlobMeta(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}
