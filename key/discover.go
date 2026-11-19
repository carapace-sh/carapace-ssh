package key

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// IdentityFile describes an SSH identity file candidate.
type IdentityFile struct {
	Path string
	Name string
}

// KeyPair groups related private key, public key, and certificate paths.
type KeyPair struct {
	Private     string
	Public      string
	Certificate string
}

// BaseName returns the key base name for the pair (private key name, or
// derived from public/certificate names when the private key is absent).
func (p KeyPair) BaseName() string {
	switch {
	case p.Private != "":
		return filepath.Base(p.Private)
	case p.Public != "":
		return strings.TrimSuffix(filepath.Base(p.Public), ".pub")
	case p.Certificate != "":
		return strings.TrimSuffix(filepath.Base(p.Certificate), "-cert.pub")
	default:
		return ""
	}
}

// DefaultDir returns the default SSH directory for the given home directory.
func DefaultDir(home string) string {
	if strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".ssh")
}

// DiscoverIdentityFiles returns private identity file candidates from dir.
// It includes regular files whose names start with "id_" and excludes public
// keys (*.pub) and certificates (*-cert.pub). Missing or unreadable
// directories are treated as non-fatal and return an empty result.
func DiscoverIdentityFiles(dir string) ([]IdentityFile, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return []IdentityFile{}, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return []IdentityFile{}, nil
		}
		return []IdentityFile{}, err
	}

	files := make([]IdentityFile, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isPrivateIdentityName(name) {
			continue
		}
		files = append(files, IdentityFile{
			Path: filepath.Join(dir, name),
			Name: name,
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return files, nil
}

// DiscoverDefaultIdentityFiles discovers identity file candidates in the default
// SSH directory for the given home directory.
func DiscoverDefaultIdentityFiles(home string) ([]IdentityFile, error) {
	return DiscoverIdentityFiles(DefaultDir(home))
}

// Paths returns the paths of the given identity files in order.
func Paths(files []IdentityFile) []string {
	if len(files) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(files))
	for _, f := range files {
		if f.Path != "" {
			out = append(out, f.Path)
		}
	}
	return out
}

// BaseNames returns deduplicated, sorted base names for discovered identity
// files (e.g. id_ed25519).
func BaseNames(files []IdentityFile) []string {
	if len(files) == 0 {
		return []string{}
	}
	seen := map[string]struct{}{}
	for _, f := range files {
		if f.Name == "" {
			continue
		}
		seen[f.Name] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// DiscoverKeyPairs discovers identity key material in dir and groups related
// private keys, public keys, and certificates into pairs. Missing or unreadable
// directories are treated as non-fatal and return an empty result.
func DiscoverKeyPairs(dir string) ([]KeyPair, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return []KeyPair{}, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return []KeyPair{}, nil
		}
		return []KeyPair{}, err
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "id_") {
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	sort.Strings(paths)
	return PairIdentityFiles(paths), nil
}

// DiscoverDefaultKeyPairs discovers and pairs identity key material in the
// default SSH directory for the given home directory.
func DiscoverDefaultKeyPairs(home string) ([]KeyPair, error) {
	return DiscoverKeyPairs(DefaultDir(home))
}

// AgentSocketPaths returns deduplicated candidate SSH agent socket paths from
// environment variables in priority order (currently SSH_AUTH_SOCK). It does
// not probe sockets, execute programs, or access the network.
func AgentSocketPaths(lookup func(string) string) []string {
	if lookup == nil {
		lookup = os.Getenv
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, 1)
	for _, key := range []string{"SSH_AUTH_SOCK"} {
		val := strings.TrimSpace(lookup(key))
		if val == "" {
			continue
		}
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		out = append(out, val)
	}
	return out
}

// DefaultAgentSocketPaths returns candidate SSH agent socket paths from the
// current process environment.
func DefaultAgentSocketPaths() []string {
	return AgentSocketPaths(os.Getenv)
}

// PublicKeyPath returns the conventional public key path for a private key path.
func PublicKeyPath(private string) string {
	if strings.TrimSpace(private) == "" {
		return ""
	}
	return private + ".pub"
}

// CertificatePath returns the conventional certificate path for a private key path.
func CertificatePath(private string) string {
	if strings.TrimSpace(private) == "" {
		return ""
	}
	return private + "-cert.pub"
}

// PairIdentityFiles groups identity-related files by private key basename within
// the same directory. Input paths may include private keys, public keys (*.pub),
// and certificates (*-cert.pub). Returned pairs are sorted by private key path,
// then public/certificate path.
func PairIdentityFiles(paths []string) []KeyPair {
	type pair struct{ private, public, cert string }
	pairs := map[string]*pair{}

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		dir, name := filepath.Split(path)
		if name == "" {
			continue
		}
		keyBase := ""
		slot := ""
		switch {
		case strings.HasSuffix(name, "-cert.pub"):
			keyBase = strings.TrimSuffix(name, "-cert.pub")
			slot = "cert"
		case strings.HasSuffix(name, ".pub"):
			keyBase = strings.TrimSuffix(name, ".pub")
			if keyBase == "" || strings.HasSuffix(keyBase, "-cert") {
				continue
			}
			slot = "public"
		default:
			keyBase = name
			slot = "private"
		}
		if keyBase == "" {
			continue
		}

		key := filepath.Clean(dir) + "\x00" + keyBase
		p := pairs[key]
		if p == nil {
			p = &pair{}
			pairs[key] = p
		}
		full := filepath.Join(dir, name)
		switch slot {
		case "cert":
			p.cert = full
		case "public":
			p.public = full
		default:
			p.private = full
		}
	}

	keys := make([]string, 0, len(pairs))
	for key := range pairs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]KeyPair, 0, len(keys))
	for _, key := range keys {
		p := pairs[key]
		out = append(out, KeyPair{
			Private:     p.private,
			Public:      p.public,
			Certificate: p.cert,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Private != out[j].Private {
			if out[i].Private == "" {
				return false
			}
			if out[j].Private == "" {
				return true
			}
			return out[i].Private < out[j].Private
		}
		pi := out[i].Public
		if pi == "" {
			pi = out[i].Certificate
		}
		pj := out[j].Public
		if pj == "" {
			pj = out[j].Certificate
		}
		return pi < pj
	})
	return out
}

func isPrivateIdentityName(name string) bool {
	if !strings.HasPrefix(name, "id_") {
		return false
	}
	if strings.HasSuffix(name, ".pub") {
		return false
	}
	return true
}
