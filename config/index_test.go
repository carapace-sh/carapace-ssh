package config

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
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustIndexFile(t *testing.T, path string, opts ...Option) *Index {
	t.Helper()
	idx, err := IndexFile(path, opts...)
	if err != nil {
		t.Fatalf("IndexFile(%s): %v", path, err)
	}
	return idx
}

func TestHostCandidatesBasicAndFiltering(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
# comment
Host foo bar *.wild ?q [abc] !negated spaced\ host
HostName ignored.example
Match host baz,*.skip,!nope
`)

	idx := mustIndexFile(t, cfg)
	got := idx.HostCandidates()
	want := []string{"bar", "baz", "foo"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
	declared := idx.DeclaredHosts()
	wantDeclared := []string{"!negated", "!nope", "*.skip", "*.wild", "?q", "[abc]", "bar", "baz", "foo", "spaced host"}
	if !reflect.DeepEqual(declared, wantDeclared) {
		t.Fatalf("DeclaredHosts() = %#v, want %#v", declared, wantDeclared)
	}
}

func TestMatchExecDoesNotExecuteAndStillIndexes(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match exec "sh -c 'echo pwned > `+marker+`'"
  Host should-still-be-indexed

Match command "touch `+marker+`" user alice
Host after-command
`)

	idx := mustIndexFile(t, cfg)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("marker file should not exist, stat err=%v", err)
	}

	got := idx.HostCandidates()
	want := []string{"after-command", "should-still-be-indexed"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestMatchOriginalHostAndCriteriaTolerance(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match originalhost bastion,*.example,!skip host prod user alice localnetwork 10.0.0.0/8 tagged deploy
Host prod-db

Match all
Host all-host

Match canonical all
Host canonical-host

Match final
Host final-host

Match localuser me localnetwork 10.0.0.0/8
Host local-host

Match user bob tagged ci
Host tagged-host

Match version 8.0 originalhost versioned
Host version-host

Match unknowncriterion value originalhost recovered
Host recovered-host

Match command "echo hi"
Host command-host

Match exec "echo hi"
Host exec-host
`)

	idx := mustIndexFile(t, cfg)
	got := idx.HostCandidates()
	want := []string{
		"all-host",
		"bastion",
		"canonical-host",
		"command-host",
		"exec-host",
		"final-host",
		"local-host",
		"prod",
		"prod-db",
		"recovered",
		"recovered-host",
		"tagged-host",
		"version-host",
		"versioned",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestHostAndMatchCommaSeparatedAndNegation(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Host one,two,*.nope,!three
Match host=alpha,beta,!gamma originalhost delta
`)

	idx := mustIndexFile(t, cfg)
	got := idx.HostCandidates()
	want := []string{"alpha", "beta", "delta", "one", "two"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIncludeResolutionAndCycle(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	main := filepath.Join(home, ".ssh", "config")
	incDir := filepath.Join(home, ".ssh", "conf.d")
	writeFile(t, filepath.Join(incDir, "b.conf"), "Host included-b\nInclude ../config\n")
	writeFile(t, filepath.Join(incDir, "a.conf"), "Host included-a\n")
	writeFile(t, filepath.Join(incDir, "space name.conf"), "Host spaced-file-host\n")
	writeFile(t, main, `
Host main-host
Include conf.d/*.conf
Include "conf.d/space name.conf"
Include missing.conf
`)

	idx := mustIndexFile(t, main, WithHomeDir(home))
	got := idx.HostCandidates()
	want := []string{"included-a", "included-b", "main-host", "spaced-file-host"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIncludeEqualsAndRelativeToIncludingFile(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "cfg")
	sub := filepath.Join(base, "sub")
	writeFile(t, filepath.Join(sub, "child.conf"), "Host child\n")
	writeFile(t, filepath.Join(base, "root.conf"), "Host root\nInclude=sub/child.conf\n")

	idx := mustIndexFile(t, filepath.Join(base, "root.conf"))
	got := idx.HostCandidates()
	want := []string{"child", "root"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIncludeTildeExpansion(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	writeFile(t, filepath.Join(home, ".ssh", "extra.conf"), "Host tilde-host\n")
	writeFile(t, filepath.Join(dir, "main.conf"), "Host main\nInclude ~/.ssh/extra.conf\n")

	idx := mustIndexFile(t, filepath.Join(dir, "main.conf"), WithHomeDir(home))
	got := idx.HostCandidates()
	want := []string{"main", "tilde-host"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestExplicitConfigDoesNotMergeDefaults(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	userCfg := filepath.Join(home, ".ssh", "config")
	sysCfg := filepath.Join(dir, "etc", "ssh", "ssh_config")
	explicit := filepath.Join(dir, "explicit.conf")

	writeFile(t, userCfg, "Host user-only\n")
	writeFile(t, sysCfg, "Host system-only\n")
	writeFile(t, explicit, "Host explicit-only\nInclude user-included.conf\n")
	writeFile(t, filepath.Join(dir, "user-included.conf"), "Host included-from-explicit\n")

	idx := mustIndexFile(t, explicit,
		WithHomeDir(home),
		WithUserConfigPath(userCfg),
		WithSystemConfigPath(sysCfg),
	)
	got := idx.HostCandidates()
	want := []string{"explicit-only", "included-from-explicit"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIndexDefaultIncludesUserAndSystem(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	userCfg := filepath.Join(home, ".ssh", "config")
	sysCfg := filepath.Join(dir, "etc", "ssh", "ssh_config")

	writeFile(t, userCfg, "Host user-only shared\n")
	writeFile(t, sysCfg, "Host system-only shared\n")

	idx, err := IndexDefault(
		WithHomeDir(home),
		WithUserConfigPath(userCfg),
		WithSystemConfigPath(sysCfg),
	)
	if err != nil {
		t.Fatalf("IndexDefault(): %v", err)
	}

	got := idx.HostCandidates()
	want := []string{"shared", "system-only", "user-only"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIndexPaths(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.conf")
	b := filepath.Join(dir, "b.conf")
	writeFile(t, a, "Host alpha shared\n")
	writeFile(t, b, "Host beta shared\n")

	idx, err := IndexPaths([]Source{{Path: a}, {Path: b}})
	if err != nil {
		t.Fatalf("IndexPaths(): %v", err)
	}
	got := idx.HostCandidates()
	want := []string{"alpha", "beta", "shared"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIndexConvenienceAndHostCandidates(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	userCfg := filepath.Join(home, ".ssh", "config")
	sysCfg := filepath.Join(dir, "etc", "ssh", "ssh_config")
	explicit := filepath.Join(dir, "explicit.conf")

	writeFile(t, userCfg, "Host user-only\n")
	writeFile(t, sysCfg, "Host system-only\n")
	writeFile(t, explicit, "Host explicit-only\n")

	got, err := HostCandidates(explicit, WithHomeDir(home), WithUserConfigPath(userCfg), WithSystemConfigPath(sysCfg))
	if err != nil {
		t.Fatalf("HostCandidates(explicit): %v", err)
	}
	if want := []string{"explicit-only"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates(explicit) = %#v, want %#v", got, want)
	}

	got, err = HostCandidates("", WithHomeDir(home), WithUserConfigPath(userCfg), WithSystemConfigPath(sysCfg))
	if err != nil {
		t.Fatalf("HostCandidates(default): %v", err)
	}
	if want := []string{"system-only", "user-only"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates(default) = %#v, want %#v", got, want)
	}

	idx, err := IndexConfig("   ", WithHomeDir(home), WithUserConfigPath(userCfg), WithSystemConfigPath(sysCfg))
	if err != nil {
		t.Fatalf("IndexConfig(whitespace): %v", err)
	}
	if want := []string{"system-only", "user-only"}; !reflect.DeepEqual(idx.HostCandidates(), want) {
		t.Fatalf("IndexConfig(whitespace).HostCandidates() = %#v, want %#v", idx.HostCandidates(), want)
	}
}

func TestSyntaxTolerance(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, "  HoSt=alpha,beta  \n\n# comment\nMatch=host\tgamma,delta\nUnknownDirective whatever\nHost epsilon\n")

	idx := mustIndexFile(t, cfg)
	got := idx.HostCandidates()
	want := []string{"alpha", "beta", "delta", "epsilon", "gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestMissingEntryIsNonFatal(t *testing.T) {
	idx, err := IndexFile(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("IndexFile(missing): %v", err)
	}
	if got := idx.HostCandidates(); len(got) != 0 {
		t.Fatalf("HostCandidates() = %#v, want empty", got)
	}
}

func TestIndexReader(t *testing.T) {
	dir := t.TempDir()
	include := filepath.Join(dir, "inc.conf")
	writeFile(t, include, "Host from-include\n")
	content := "Host from-reader\nInclude inc.conf\nMatch exec \"echo should-not-run\"\nHost safe\n"

	idx, err := IndexReader(strings.NewReader(content), Source{Path: filepath.Join(dir, "stdin.conf"), BaseDir: dir})
	if err != nil {
		t.Fatalf("IndexReader(): %v", err)
	}
	got := idx.HostCandidates()
	want := []string{"from-include", "from-reader", "safe"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestFixtureConfig(t *testing.T) {
	idx := mustIndexFile(t, filepath.Join("testdata", "hosts.conf"))
	got := idx.HostCandidates()
	want := []string{"all-host", "alpha", "beta", "delta", "gamma", "gated-by-exec", "jump"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestMatchCriteriaWithSpacedEquals(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match host = alpha originalhost= beta user = alice
Host from-spaced-equals
`)

	idx := mustIndexFile(t, cfg)
	got := idx.HostCandidates()
	want := []string{"alpha", "beta", "from-spaced-equals"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestLineContinuation(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, "Host alpha \\\n  beta\nMatch host gamma \\\n  originalhost delta\n")

	idx := mustIndexFile(t, cfg)
	got := idx.HostCandidates()
	want := []string{"alpha", "beta", "delta", "gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}
