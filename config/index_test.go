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

func TestSystemConfigDropInIncludes(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	sysCfg := filepath.Join(dir, "etc", "ssh", "ssh_config")
	dropIn := filepath.Join(dir, "etc", "ssh", "ssh_config.d", "10-example.conf")
	writeFile(t, dropIn, "Host dropin-host\n")
	writeFile(t, sysCfg, "Host system-host\nInclude ssh_config.d/*.conf\n")

	idx, err := IndexDefault(
		WithHomeDir(home),
		WithUserConfigPath(filepath.Join(home, ".ssh", "config")),
		WithSystemConfigPath(sysCfg),
	)
	if err != nil {
		t.Fatalf("IndexDefault(): %v", err)
	}
	got := idx.HostCandidates()
	want := []string{"dropin-host", "system-host"}
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

func TestIndexPathsSkipsEmptySourcePath(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	a := filepath.Join(dir, "a.conf")
	writeFile(t, a, "Host alpha\n")

	idx, err := IndexPaths([]Source{{}, {Path: a}}, WithHomeDir(home))
	if err != nil {
		t.Fatalf("IndexPaths(): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"alpha"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestDefaultSources(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	userCfg := filepath.Join(home, ".ssh", "config")
	sysCfg := filepath.Join(dir, "etc", "ssh", "ssh_config")

	sources := DefaultSources(WithHomeDir(home), WithUserConfigPath(userCfg), WithSystemConfigPath(sysCfg))
	want := []Source{
		{Path: userCfg, BaseDir: filepath.Join(home, ".ssh"), System: false},
		{Path: sysCfg, BaseDir: filepath.Join(dir, "etc", "ssh"), System: true},
	}
	if !reflect.DeepEqual(sources, want) {
		t.Fatalf("DefaultSources() = %#v, want %#v", sources, want)
	}

	sources = DefaultSources(WithHomeDir(home), WithUserConfigPath(""), WithSystemConfigPath(sysCfg))
	want = []Source{{Path: sysCfg, BaseDir: filepath.Join(dir, "etc", "ssh"), System: true}}
	if !reflect.DeepEqual(sources, want) {
		t.Fatalf("DefaultSources(user empty) = %#v, want %#v", sources, want)
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

func TestStaticFacts(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Host foo bar
  HostName foo.internal
  User alice
  Port 2222
  IdentityFile ~/.ssh/id_foo
  IdentityFile ~/.ssh/id_foo2
  CertificateFile ~/.ssh/foo-cert.pub
  ProxyJump jump.example
  ProxyCommand ssh -W %h:%p jump.example
  IdentityAgent ~/.ssh/agent.sock
  UserKnownHostsFile ~/.ssh/known_hosts.custom ~/.ssh/known_hosts.shared
  GlobalKnownHostsFile /etc/ssh/known_hosts.custom

Host baz
  HostName baz.internal
  GlobalKnownHostsFile /etc/ssh/known_hosts.shared /etc/ssh/known_hosts.custom

Match host prod originalhost=prod.example user deploy
  User produser
  Port 2200

Match exec "echo hi"
  Host gated
  User gateduser
`)

	idx := mustIndexFile(t, cfg)

	f := idx.Facts("foo")
	if f.HostName != "foo.internal" || f.User != "alice" || f.Port != "2222" {
		t.Fatalf("foo basic facts = %#v", f)
	}
	if !reflect.DeepEqual(f.IdentityFiles, []string{"~/.ssh/id_foo", "~/.ssh/id_foo2"}) {
		t.Fatalf("foo IdentityFiles = %#v", f.IdentityFiles)
	}
	if !reflect.DeepEqual(f.CertificateFiles, []string{"~/.ssh/foo-cert.pub"}) {
		t.Fatalf("foo CertificateFiles = %#v", f.CertificateFiles)
	}
	if f.ProxyJump != "jump.example" {
		t.Fatalf("foo ProxyJump = %#v", f.ProxyJump)
	}
	if f.ProxyCommand != "ssh -W %h:%p jump.example" {
		t.Fatalf("foo ProxyCommand = %#v", f.ProxyCommand)
	}
	if f.IdentityAgent != "~/.ssh/agent.sock" {
		t.Fatalf("foo IdentityAgent = %#v", f.IdentityAgent)
	}
	if !reflect.DeepEqual(f.UserKnownHostsFiles, []string{"~/.ssh/known_hosts.custom", "~/.ssh/known_hosts.shared"}) {
		t.Fatalf("foo UserKnownHostsFiles = %#v", f.UserKnownHostsFiles)
	}
	if !reflect.DeepEqual(f.GlobalKnownHostsFiles, []string{"/etc/ssh/known_hosts.custom"}) {
		t.Fatalf("foo GlobalKnownHostsFiles = %#v", f.GlobalKnownHostsFiles)
	}

	if bf := idx.Facts("bar"); bf.HostName != "foo.internal" || bf.User != "alice" {
		t.Fatalf("bar facts = %#v", bf)
	}
	if pf := idx.Facts("prod"); pf.User != "produser" || pf.Port != "2200" {
		t.Fatalf("prod facts = %#v", pf)
	}
	if pf := idx.Facts("prod.example"); pf.User != "produser" {
		t.Fatalf("prod.example facts = %#v", pf)
	}
	if gf := idx.Facts("gated"); gf.User != "gateduser" {
		t.Fatalf("gated facts = %#v", gf)
	}

	hosts := idx.HostsWithFacts()
	wantHosts := []string{"bar", "baz", "foo", "gated", "prod", "prod.example"}
	if !reflect.DeepEqual(hosts, wantHosts) {
		t.Fatalf("HostsWithFacts() = %#v, want %#v", hosts, wantHosts)
	}

	if got, want := idx.KnownHostsFiles(), []string{"/etc/ssh/known_hosts.custom", "/etc/ssh/known_hosts.shared", "~/.ssh/known_hosts.custom", "~/.ssh/known_hosts.shared"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("KnownHostsFiles() = %#v, want %#v", got, want)
	}

	if facts := idx.Facts("unknown"); !reflect.DeepEqual(facts, HostFacts{}) {
		t.Fatalf("Facts(unknown) = %#v, want zero", facts)
	}
}

func TestStaticFactsAgentCandidates(t *testing.T) {
	t.Setenv("SSH_AUTH_SOCK", "/tmp/test-agent.sock")
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Host h1
  IdentityAgent none
Host h2
  IdentityAgent SSH_AUTH_SOCK
Host h3
  IdentityAgent /custom/agent.sock
`)

	idx := mustIndexFile(t, cfg)
	if got := idx.Facts("h1").AgentCandidates(); got != nil {
		t.Fatalf("h1 AgentCandidates() = %#v, want nil", got)
	}
	if got := idx.Facts("h2").AgentCandidates(); !reflect.DeepEqual(got, []string{"/tmp/test-agent.sock"}) {
		t.Fatalf("h2 AgentCandidates() = %#v, want [/tmp/test-agent.sock]", got)
	}
	if got := idx.Facts("h3").AgentCandidates(); !reflect.DeepEqual(got, []string{"/custom/agent.sock"}) {
		t.Fatalf("h3 AgentCandidates() = %#v, want [/custom/agent.sock]", got)
	}
}

func TestUnreadableEntryIsNonFatal(t *testing.T) {
	idx, err := IndexFile("/unreadable/config", WithReadFile(func(string) ([]byte, error) {
		return nil, os.ErrPermission
	}))
	if err != nil {
		t.Fatalf("IndexFile(unreadable): %v", err)
	}
	if got := idx.HostCandidates(); len(got) != 0 {
		t.Fatalf("HostCandidates() = %#v, want empty", got)
	}

	idx, err = IndexDefault(
		WithUserConfigPath("/unreadable/user"),
		WithSystemConfigPath("/unreadable/system"),
		WithReadFile(func(string) ([]byte, error) {
			return nil, os.ErrPermission
		}),
	)
	if err != nil {
		t.Fatalf("IndexDefault(unreadable): %v", err)
	}
	if got := idx.HostCandidates(); len(got) != 0 {
		t.Fatalf("IndexDefault HostCandidates() = %#v, want empty", got)
	}
}
