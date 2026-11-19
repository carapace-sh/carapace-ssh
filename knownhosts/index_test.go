package knownhosts

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestIndexFileParsesKnownHosts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known_hosts")
	writeFile(t, path, `
# comment
github.com,140.82.121.4 ssh-ed25519 AAAA
[example.com]:2222,[10.0.0.1]:2222 ssh-rsa AAAA
[2001:db8::1]:2222 ssh-ed25519 AAAA
@cert-authority *.trusted.example,ca.example ssh-rsa AAAA
@revoked revoked.example,also-revoked.example ssh-rsa AAAA
|1|hashsalt|hashvalue ssh-rsa AAAA
[bad host]:22 ssh-rsa AAAA
`)

	idx, err := IndexFile(path)
	if err != nil {
		t.Fatalf("IndexFile(): %v", err)
	}
	got := idx.HostCandidates()
	// @revoked entries must not be suggested; @cert-authority may contribute literal hosts.
	// Bracketed IPv6 addresses are extracted as raw IPv6 literals.
	want := []string{"10.0.0.1", "140.82.121.4", "2001:db8::1", "ca.example", "example.com", "github.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIndexPathsAndDefault(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	user := filepath.Join(home, ".ssh", "known_hosts")
	sys := filepath.Join(dir, "etc", "ssh", "ssh_known_hosts")
	writeFile(t, user, "alpha.example ssh-ed25519 AAAA\nshared.example ssh-ed25519 AAAA\n")
	writeFile(t, sys, "beta.example ssh-ed25519 AAAA\nshared.example ssh-ed25519 AAAA\n")

	idx, err := IndexPaths([]string{user, sys})
	if err != nil {
		t.Fatalf("IndexPaths(): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"alpha.example", "beta.example", "shared.example"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("IndexPaths HostCandidates() = %#v, want %#v", got, want)
	}

	idx, err = IndexDefault(WithHomeDir(home), WithUserKnownHostsPath(user), WithSystemKnownHostsPath(sys))
	if err != nil {
		t.Fatalf("IndexDefault(): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"alpha.example", "beta.example", "shared.example"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("IndexDefault HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIndexReaderAndMerge(t *testing.T) {
	idx, err := IndexReader(strings.NewReader("one.example ssh-ed25519 AAAA\n[two.example]:22 ssh-ed25519 AAAA\n"))
	if err != nil {
		t.Fatalf("IndexReader(): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"one.example", "two.example"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}

	merged := Merge([]string{"b", "a"}, []string{"a", "c"}, nil)
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(merged, want) {
		t.Fatalf("Merge() = %#v, want %#v", merged, want)
	}
}

func TestDefaultPaths(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	user := filepath.Join(home, ".ssh", "known_hosts")
	sys := filepath.Join(dir, "etc", "ssh", "ssh_known_hosts")

	got := DefaultPaths(WithHomeDir(home), WithSystemKnownHostsPath(sys))
	want := []string{user, sys}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DefaultPaths() = %#v, want %#v", got, want)
	}

	got = DefaultPaths(WithUserKnownHostsPath("/custom/user_known_hosts"), WithSystemKnownHostsPath(""))
	want = []string{"/custom/user_known_hosts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DefaultPaths(custom) = %#v, want %#v", got, want)
	}
}

func TestMissingKnownHostsIsNonFatal(t *testing.T) {
	idx, err := IndexFile(filepath.Join(t.TempDir(), "missing-known_hosts"))
	if err != nil {
		t.Fatalf("IndexFile(missing): %v", err)
	}
	if got := idx.HostCandidates(); len(got) != 0 {
		t.Fatalf("HostCandidates() = %#v, want empty", got)
	}
}

func TestIndexDefaultUsesHomeDir(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	user := filepath.Join(home, ".ssh", "known_hosts")
	sys := filepath.Join(dir, "etc", "ssh", "ssh_known_hosts")
	writeFile(t, user, "home-user.example ssh-ed25519 AAAA\n")
	writeFile(t, sys, "home-system.example ssh-ed25519 AAAA\n")

	idx, err := IndexDefault(WithHomeDir(home), WithSystemKnownHostsPath(sys))
	if err != nil {
		t.Fatalf("IndexDefault(): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"home-system.example", "home-user.example"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestUnreadableKnownHostsIsNonFatal(t *testing.T) {
	idx, err := IndexFile("/unreadable/known_hosts", WithReadFile(func(string) ([]byte, error) {
		return nil, os.ErrPermission
	}))
	if err != nil {
		t.Fatalf("IndexFile(unreadable): %v", err)
	}
	if got := idx.HostCandidates(); len(got) != 0 {
		t.Fatalf("HostCandidates() = %#v, want empty", got)
	}

	idx, err = IndexPaths([]string{"/unreadable/known_hosts"}, WithReadFile(func(string) ([]byte, error) {
		return nil, os.ErrPermission
	}))
	if err != nil {
		t.Fatalf("IndexPaths(unreadable): %v", err)
	}
	if got := idx.HostCandidates(); len(got) != 0 {
		t.Fatalf("HostCandidates() = %#v, want empty", got)
	}
}

func TestMarkerAndMixedHostTokenPolicies(t *testing.T) {
	idx, err := IndexReader(strings.NewReader(`
@revoked revoked-only.example ssh-ed25519 AAAA
@revoked revoked.example,still-revoked.example ssh-rsa AAAA
@cert-authority *.wild.example ssh-rsa AAAA
@cert-authority ca-literal.example,*.wild.example ssh-rsa AAAA
mixed.example,*.wild.example,!neg.example,[2001:db8::2]:22 ssh-ed25519 AAAA
|1|hash|hash ssh-rsa AAAA
`))
	if err != nil {
		t.Fatalf("IndexReader(): %v", err)
	}
	// Revoked entries are never suggested. CA entries may contribute explicit
	// literal hosts, but wildcard-only scopes do not. Mixed lists keep only
	// literal, non-negated, non-wildcard, non-hashed hosts (IPv6 extracted raw).
	got := idx.HostCandidates()
	want := []string{"2001:db8::2", "ca-literal.example", "mixed.example"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestKnownHostsFixture(t *testing.T) {
	idx, err := IndexFile(filepath.Join("testdata", "hosts-known_hosts"))
	if err != nil {
		t.Fatalf("IndexFile(): %v", err)
	}
	// @revoked excluded, hashed skipped, wildcard/negated filtered; CA literal
	// hosts may be included when explicitly present. Bracketed IPv6 extracted raw.
	got := idx.HostCandidates()
	want := []string{"140.82.121.4", "2001:db8::1", "ca-literal.example", "example.com", "github.com", "mixed.example"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestKnownHostsBracketedHostWithoutPortAndMalformed(t *testing.T) {
	idx, err := IndexReader(strings.NewReader(`
[example.com] ssh-ed25519 AAAA
[2001:db8::5] ssh-ed25519 AAAA
[broken ssh-ed25519 AAAA
[]:22 ssh-ed25519 AAAA
just-a-host
`))
	if err != nil {
		t.Fatalf("IndexReader(): %v", err)
	}
	// Bracketed hosts without :port are accepted as host tokens; malformed
	// bracket forms and bare key-only lines should not produce candidates.
	got := idx.HostCandidates()
	want := []string{"2001:db8::5", "example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}
