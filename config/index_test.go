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

Match sessiontype shell originalhost session-host
Host session-declared

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
		"session-declared",
		"session-host",
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

func TestIndexPathsPreservesSystemSourceContext(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	sysBase := filepath.Join(dir, "etc", "ssh")
	sysCfg := filepath.Join(sysBase, "ssh_config")
	writeFile(t, filepath.Join(sysBase, "included.conf"), "Host from-system-include\n")
	writeFile(t, sysCfg, "Host system-main\nInclude included.conf\n")

	idx, err := IndexPaths([]Source{{Path: sysCfg, System: true}}, WithHomeDir(home), WithSystemConfigPath(sysCfg))
	if err != nil {
		t.Fatalf("IndexPaths(): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"from-system-include", "system-main"}; !reflect.DeepEqual(got, want) {
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

func TestMatchEmptyCriteriaArgs(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match tagged "" host from-empty-tagged
Host tagged-host

Match command "" host from-empty-command
Host command-host
`)

	idx := mustIndexFile(t, cfg)
	got := idx.HostCandidates()
	want := []string{"command-host", "from-empty-command", "from-empty-tagged", "tagged-host"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIncludeInsideHostBlockFollowed(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "cfg")
	writeFile(t, filepath.Join(base, "inc.conf"), "Host included-via-host-block\n")
	writeFile(t, filepath.Join(base, "main.conf"), `
Host wrapper
  Include inc.conf
  HostName wrapper.internal

Host outside
`)
	idx := mustIndexFile(t, filepath.Join(base, "main.conf"))
	got := idx.HostCandidates()
	want := []string{"included-via-host-block", "outside", "wrapper"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
	if f := idx.Facts("wrapper"); f.HostName != "wrapper.internal" {
		t.Fatalf("Facts(wrapper).HostName = %#v, want %q", f.HostName, "wrapper.internal")
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
  GlobalKnownHostsFile /etc/ssh/known_hosts.shared none /etc/ssh/known_hosts.custom

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

func TestKnownHostsFilesExcludesNone(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Host a
  UserKnownHostsFile none ~/.ssh/known_hosts.a
  GlobalKnownHostsFile /etc/ssh/kh_global NONE

Host b
  UserKnownHostsFile ~/.ssh/known_hosts.b
`)

	idx := mustIndexFile(t, cfg)
	if got, want := idx.KnownHostsFiles(), []string{"/etc/ssh/kh_global", "~/.ssh/known_hosts.a", "~/.ssh/known_hosts.b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("KnownHostsFiles() = %#v, want %#v", got, want)
	}
	// Declared facts preserve values as written.
	if got := idx.Facts("a").UserKnownHostsFiles; !reflect.DeepEqual(got, []string{"none", "~/.ssh/known_hosts.a"}) {
		t.Fatalf("Facts(a).UserKnownHostsFiles = %#v", got)
	}
}

func TestMatchExecInlineAndQuotedFormsDoNotExecute(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match exec="sh -c 'echo x > `+marker+`'"
Host after-inline-exec

Match exec "sh -c \"echo x > `+marker+`\"" originalhost quoted-exec-host
Host after-quoted-exec
`)
	// Note: exec consumes its command argument opaquely; hosts in the same Match
	// line after exec may not be recoverable without execution semantics. Safety
	// and continued indexing are the required guarantees.

	idx := mustIndexFile(t, cfg)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("marker file should not exist, stat err=%v", err)
	}
	got := idx.HostCandidates()
	want := []string{"after-inline-exec", "after-quoted-exec"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestCRLFAndContinuationTolerance(t *testing.T) {
	idx, err := IndexReader(strings.NewReader("Host alpha\r\nMatch host beta \\\r\n  originalhost gamma\r\nHost delta\r\n"), Source{Path: "crlf.conf", BaseDir: "."})
	if err != nil {
		t.Fatalf("IndexReader(): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"alpha", "beta", "delta", "gamma"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestCommentsAndQuotingDoNotBreakParsing(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
# leading comment
Host keep # trailing comment
  HostName kept.example # value comment

Host "quoted-host" # quoted token
  ProxyCommand ssh -W %h:%p "jump host" # quoted arg with space
`)
	idx := mustIndexFile(t, cfg)
	if got, want := idx.HostCandidates(), []string{"keep", "quoted-host"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
	if got, want := idx.DeclaredHosts(), []string{"keep", "quoted-host"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("DeclaredHosts() = %#v, want %#v", got, want)
	}
	if f := idx.Facts("keep"); f.HostName != "kept.example" {
		t.Fatalf("Facts(keep).HostName = %#v, want %q", f.HostName, "kept.example")
	}
}

func TestIndexReaderIncludeBaseDirContexts(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "inc.conf"), "Host from-inc\n")
	writeFile(t, filepath.Join(dir, "system-inc.conf"), "Host from-system-inc\n")

	// Relative include resolved using BaseDir.
	idx, err := IndexReader(strings.NewReader("Host root\nInclude inc.conf\n"), Source{Path: filepath.Join(dir, "main.conf"), BaseDir: dir})
	if err != nil {
		t.Fatalf("IndexReader(BaseDir): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"from-inc", "root"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates(BaseDir) = %#v, want %#v", got, want)
	}

	// System context with BaseDir resolves relative includes in that base.
	idx, err = IndexReader(strings.NewReader("Host sysroot\nInclude system-inc.conf\n"), Source{Path: filepath.Join(dir, "ssh_config"), BaseDir: dir, System: true})
	if err != nil {
		t.Fatalf("IndexReader(system): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"from-system-inc", "sysroot"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates(system) = %#v, want %#v", got, want)
	}
}

func TestMalformedAndBinaryInputIsTolerated(t *testing.T) {
	idx, err := IndexReader(strings.NewReader("Host ok\n\x00\xff\xfe not-a-directive\nMatch host still-ok\n"), Source{Path: "bin.conf", BaseDir: "."})
	if err != nil {
		t.Fatalf("IndexReader(): %v", err)
	}
	// Best-effort parsing: must not panic and should recover valid directives.
	got := idx.HostCandidates()
	want := []string{"ok", "still-ok"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestMatchNegatedHostCriteriaDoNotExtractHosts(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match !host bad.example host good.example
Host declared-after

Match !originalhost nope.example originalhost keep.example
Host declared-after-2
`)
	idx := mustIndexFile(t, cfg)
	if got, want := idx.HostCandidates(), []string{"declared-after", "declared-after-2", "good.example", "keep.example"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
	declared := idx.DeclaredHosts()
	for _, bad := range []string{"bad.example", "nope.example"} {
		for _, h := range declared {
			if h == bad {
				t.Fatalf("DeclaredHosts() unexpectedly contains %q: %#v", bad, declared)
			}
		}
	}
}

func TestMatchNegatedHostCriteriaInlineAndMixed(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match !host=bad.example host=good.example
Host declared-after

Match user=alice !originalhost skip.example originalhost keep.example tagged ""
Host declared-after-2
`)
	idx := mustIndexFile(t, cfg)
	if got, want := idx.HostCandidates(), []string{"declared-after", "declared-after-2", "good.example", "keep.example"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
	declared := idx.DeclaredHosts()
	for _, bad := range []string{"bad.example", "skip.example"} {
		for _, h := range declared {
			if h == bad {
				t.Fatalf("DeclaredHosts() unexpectedly contains %q: %#v", bad, declared)
			}
		}
	}
}

func TestEdgeCaseFixture(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "exec-marker")
	content, err := os.ReadFile(filepath.Join("testdata", "edge-cases.conf"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	cfg := filepath.Join(t.TempDir(), "edge-cases.conf")
	replaced := strings.ReplaceAll(string(content), "/tmp/should-not-exist-carapace-ssh", marker)
	writeFile(t, cfg, replaced)

	// Ensure fixture includes are resolved relative to the temp config location.
	writeFile(t, filepath.Join(filepath.Dir(cfg), "includes", "extra.conf"), "Host gamma\nMatch host delta,!*.skip\n  User someone\n")
	writeFile(t, filepath.Join(filepath.Dir(cfg), "edge-includes", "inc.conf"), "Host included-one\n  IdentityFile ~/.ssh/id_included\n")

	idx := mustIndexFile(t, cfg)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("exec marker must not exist, stat err=%v", err)
	}

	gotHosts := idx.HostCandidates()
	wantHosts := []string{"alpha", "beta", "cmd-empty", "delta", "gamma", "gated-by-exec", "good.example", "included-one", "orig.example", "prod"}
	if !reflect.DeepEqual(gotHosts, wantHosts) {
		t.Fatalf("HostCandidates() = %#v, want %#v", gotHosts, wantHosts)
	}

	declared := idx.DeclaredHosts()
	for _, must := range []string{"alpha", "beta", "cmd-empty", "delta", "gamma", "gated-by-exec", "good.example", "included-one", "orig.example", "prod", "*.wild", "!negated", "*.skip", "!*.skip"} {
		found := false
		for _, h := range declared {
			if h == must {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("DeclaredHosts() missing %q: %#v", must, declared)
		}
	}
	for _, bad := range []string{"bad.example"} {
		for _, h := range declared {
			if h == bad {
				t.Fatalf("DeclaredHosts() unexpectedly contains %q: %#v", bad, declared)
			}
		}
	}

	if f := idx.Facts("alpha"); f.HostName != "alpha.internal" || f.User != "alice" || f.Port != "2222" {
		t.Fatalf("Facts(alpha) = %#v", f)
	} else if !reflect.DeepEqual(f.UserKnownHostsFiles, []string{"none", "~/.ssh/kh_alpha"}) {
		t.Fatalf("Facts(alpha).UserKnownHostsFiles = %#v", f.UserKnownHostsFiles)
	}

	if f := idx.Facts("prod"); f.ProxyJump != "bastion.example" || !reflect.DeepEqual(f.GlobalKnownHostsFiles, []string{"/etc/ssh/kh_prod"}) {
		t.Fatalf("Facts(prod) = %#v", f)
	}
	if f := idx.Facts("good.example"); f.User != "gooduser" {
		t.Fatalf("Facts(good.example).User = %#v", f.User)
	}
	if f := idx.Facts("gated-by-exec"); f.User != "gated" {
		t.Fatalf("Facts(gated-by-exec).User = %#v", f.User)
	}
	if f := idx.Facts("cmd-empty"); f.HostName != "cmd.example" {
		t.Fatalf("Facts(cmd-empty).HostName = %#v", f.HostName)
	}
	if f := idx.Facts("included-one"); !reflect.DeepEqual(f.IdentityFiles, []string{"~/.ssh/id_included"}) {
		t.Fatalf("Facts(included-one).IdentityFiles = %#v", f.IdentityFiles)
	}
	if f := idx.Facts("delta"); f.User != "someone" {
		t.Fatalf("Facts(delta).User = %#v", f.User)
	}

	if got, want := idx.KnownHostsFiles(), []string{"/etc/ssh/kh_prod", "~/.ssh/kh_alpha"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("KnownHostsFiles() = %#v, want %#v", got, want)
	}
}

func TestMatchNegatedNonHostCriteriaStillExtractHosts(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match !user root host ok1
Host h1

Match host ok2 !localnetwork 10.0.0.0/8
Host h2

Match !tagged bad host ok3 tagged ""
Host h3

Match !sessiontype none originalhost ok4
Host h4
`)
	idx := mustIndexFile(t, cfg)
	if got, want := idx.HostCandidates(), []string{"h1", "h2", "h3", "h4", "ok1", "ok2", "ok3", "ok4"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestMatchCriteriaAlignmentEdgeCases(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match canonical all host canon-host
Host h-canon

Match final host final-host originalhost final-orig
Host h-final

Match user alice host first originalhost second tagged ok
Host h-multi

Match host = spaced-host user = bob originalhost=spaced-orig
Host h-spaced
`)
	idx := mustIndexFile(t, cfg)
	if got, want := idx.HostCandidates(), []string{"canon-host", "final-host", "final-orig", "first", "h-canon", "h-final", "h-multi", "h-spaced", "second", "spaced-host", "spaced-orig"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestMatchNegatedCommandAndExecSafety(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match !command sftp host after-neg-command
Host neg-command-target

Match !exec "sh -c 'echo x > `+marker+`'"
Host after-neg-exec

Match exec="sh -c 'echo x > `+marker+`'"
Host after-inline-exec-eq
`)
	idx := mustIndexFile(t, cfg)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("marker file should not exist, stat err=%v", err)
	}
	// !command consumes a value argument and host extraction may continue.
	// !exec/exec consume opaque command text for safety; later criteria on the
	// same line may not be recoverable, but indexing must continue safely.
	got := idx.HostCandidates()
	want := []string{"after-inline-exec-eq", "after-neg-command", "after-neg-exec", "neg-command-target"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestMatchCommandInlineAndCanonicalFinalForms(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match command=sftp host sftp-host
Host h1

Match canonical
Host canon-block

Match final
Host final-block

Match canonical host canon-host2
Host h2
`)
	idx := mustIndexFile(t, cfg)
	got := idx.HostCandidates()
	want := []string{"canon-block", "canon-host2", "final-block", "h1", "h2", "sftp-host"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestMatchUnknownCriterionFallbackPreservesKnownHosts(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match unknowncriterion host ok1
Host h1

Match unknowncriterion=value host ok2 originalhost ok3
Host h2

Match unknowncriterion = value host ok4
Host h3

Match unknowncriterion unknownvalue host ok5
Host h4
`)
	idx := mustIndexFile(t, cfg)
	if got, want := idx.HostCandidates(), []string{"h1", "h2", "h3", "h4", "ok1", "ok2", "ok3", "ok4", "ok5"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}

func TestIndexReaderDefaultBaseDirFallback(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	writeFile(t, filepath.Join(home, ".ssh", "inc.conf"), "Host from-home-ssh\n")
	writeFile(t, filepath.Join(dir, "etc", "ssh", "inc.conf"), "Host from-etc-ssh\n")

	// No Path/BaseDir: user context falls back to ~/.ssh for relative includes.
	idx, err := IndexReader(strings.NewReader("Host root\nInclude inc.conf\n"), Source{}, WithHomeDir(home))
	if err != nil {
		t.Fatalf("IndexReader(user fallback): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"from-home-ssh", "root"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates(user fallback) = %#v, want %#v", got, want)
	}

	// No Path/BaseDir, System context falls back to system config directory.
	sysCfg := filepath.Join(dir, "etc", "ssh", "ssh_config")
	idx, err = IndexReader(strings.NewReader("Host sysroot\nInclude inc.conf\n"), Source{System: true}, WithHomeDir(home), WithSystemConfigPath(sysCfg))
	if err != nil {
		t.Fatalf("IndexReader(system fallback): %v", err)
	}
	if got, want := idx.HostCandidates(), []string{"from-etc-ssh", "sysroot"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates(system fallback) = %#v, want %#v", got, want)
	}
}

func TestUnexpandedTokensAndEnvFormsAreTolerated(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Host tok
  HostName %h.example
  User %r
  IdentityFile ~/.ssh/id_%h
  ProxyCommand ssh -W %h:%p $JUMP_HOST

Include config.d/%h.conf
Include $UNEXPANDED_CONF
`)
	// Must not crash; include targets with unexpanded tokens/env are non-fatal.
	idx := mustIndexFile(t, cfg)
	if got, want := idx.HostCandidates(), []string{"tok"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
	f := idx.Facts("tok")
	if f.HostName != "%h.example" || f.User != "%r" {
		t.Fatalf("Facts(tok) basic = %#v", f)
	}
	if !reflect.DeepEqual(f.IdentityFiles, []string{"~/.ssh/id_%h"}) {
		t.Fatalf("Facts(tok).IdentityFiles = %#v", f.IdentityFiles)
	}
	if f.ProxyCommand != "ssh -W %h:%p $JUMP_HOST" {
		t.Fatalf("Facts(tok).ProxyCommand = %#v", f.ProxyCommand)
	}
}

func TestMatchHostListNegatedValuesDoNotBecomeCompletable(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match host good,!bad,*.skip
  User alice
`)
	idx := mustIndexFile(t, cfg)
	if got, want := idx.HostCandidates(), []string{"good"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
	if got, want := idx.DeclaredHosts(), []string{"!bad", "*.skip", "good"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("DeclaredHosts() = %#v, want %#v", got, want)
	}
	// Facts are associated to declared block host tokens (static/best-effort).
	if f := idx.Facts("good"); f.User != "alice" {
		t.Fatalf("Facts(good).User = %#v, want %q", f.User, "alice")
	}
	if f := idx.Facts("!bad"); f.User != "alice" {
		t.Fatalf("Facts(!bad).User = %#v, want %q", f.User, "alice")
	}
}

func TestMatchHostExtractionOrderWithExec(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")
	cfg := filepath.Join(dir, "config")
	writeFile(t, cfg, `
Match host before-exec exec "sh -c 'echo x > `+marker+`'"
Host after-1

Match exec "sh -c 'echo x > `+marker+`'" host after-exec
Host after-2
`)
	idx := mustIndexFile(t, cfg)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("marker file should not exist, stat err=%v", err)
	}
	// host extracted before exec is recoverable; host after exec may not be due to
	// opaque command-text safety (must not execute or misparse command as criteria).
	got := idx.HostCandidates()
	want := []string{"after-1", "after-2", "before-exec"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
}
