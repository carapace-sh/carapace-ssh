package authorizedkeys

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
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		ok      bool
		options []string
		keyType string
		key     string
		comment string
	}{
		{name: "simple", line: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest alice@example", ok: true, keyType: "ssh-ed25519", key: "AAAAC3NzaC1lZDI1NTE5AAAAITest", comment: "alice@example"},
		{name: "rsa no comment", line: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQ==", ok: true, keyType: "ssh-rsa", key: "AAAAB3NzaC1yc2EAAAADAQABAAABgQ=="},
		{name: "with options", line: `command="echo hi",no-pty ssh-ed25519 AAAA key1`, ok: true, options: []string{`command="echo hi"`, "no-pty"}, keyType: "ssh-ed25519", key: "AAAA", comment: "key1"},
		{name: "quoted comma option", line: `from="*.example.com,10.0.0.0/8",restrict ssh-rsa AAAA user`, ok: true, options: []string{`from="*.example.com,10.0.0.0/8"`, "restrict"}, keyType: "ssh-rsa", key: "AAAA", comment: "user"},
		{name: "ecdsa", line: "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY= comment words", ok: true, keyType: "ecdsa-sha2-nistp256", key: "AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY=", comment: "comment words"},
		{name: "cert keytype", line: "ssh-ed25519-cert-v01@openssh.com AAAA cert", ok: true, keyType: "ssh-ed25519-cert-v01@openssh.com", key: "AAAA", comment: "cert"},
		{name: "sk keytype", line: "sk-ssh-ed25519@openssh.com AAAA security key", ok: true, keyType: "sk-ssh-ed25519@openssh.com", key: "AAAA", comment: "security key"},
		{name: "empty", line: "", ok: false},
		{name: "comment", line: "# comment", ok: false},
		{name: "invalid", line: "not-a-key", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := ParseLine(tt.line)
			if ok != tt.ok {
				t.Fatalf("ParseLine(%q) ok=%v, want %v (entry=%#v)", tt.line, ok, tt.ok, entry)
			}
			if !ok {
				return
			}
			if !reflect.DeepEqual(entry.Options, tt.options) {
				t.Fatalf("Options = %#v, want %#v", entry.Options, tt.options)
			}
			if entry.KeyType != tt.keyType {
				t.Fatalf("KeyType = %q, want %q", entry.KeyType, tt.keyType)
			}
			if entry.Key != tt.key {
				t.Fatalf("Key = %q, want %q", entry.Key, tt.key)
			}
			if entry.Comment != tt.comment {
				t.Fatalf("Comment = %q, want %q", entry.Comment, tt.comment)
			}
		})
	}
}

func TestIndexReaderAndComments(t *testing.T) {
	content := strings.Join([]string{
		"# ignore",
		"",
		"ssh-ed25519 AAAA alice",
		`command="uptime" ssh-rsa BBBB bob`,
		"invalid-line",
		"ecdsa-sha2-nistp256 CCCC",
	}, "\n")
	idx, err := IndexReader(strings.NewReader(content))
	if err != nil {
		t.Fatalf("IndexReader(): %v", err)
	}
	entries := idx.Entries()
	if len(entries) != 3 {
		t.Fatalf("Entries() len=%d, want 3 (%#v)", len(entries), entries)
	}
	if got, want := idx.Comments(), []string{"alice", "bob"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Comments() = %#v, want %#v", got, want)
	}
	if entries[1].KeyType != "ssh-rsa" || !reflect.DeepEqual(entries[1].Options, []string{`command="uptime"`}) {
		t.Fatalf("entry[1] = %#v", entries[1])
	}
}

func TestIndexFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "authorized_keys")
	writeFile(t, path, "ssh-ed25519 AAAA test@host\n")

	idx, err := IndexFile(path)
	if err != nil {
		t.Fatalf("IndexFile(): %v", err)
	}
	if len(idx.Entries()) != 1 {
		t.Fatalf("Entries() = %#v, want 1 entry", idx.Entries())
	}

	idx, err = IndexFile(filepath.Join(dir, "missing"))
	if err != nil {
		t.Fatalf("IndexFile(missing): %v", err)
	}
	if len(idx.Entries()) != 0 {
		t.Fatalf("IndexFile(missing) entries = %#v, want empty", idx.Entries())
	}

	idx, err = IndexFile("")
	if err != nil {
		t.Fatalf("IndexFile(empty): %v", err)
	}
	if len(idx.Entries()) != 0 {
		t.Fatalf("IndexFile(empty) entries = %#v, want empty", idx.Entries())
	}
}

func TestDefaultPathAndIndexDefault(t *testing.T) {
	if got := DefaultPath(""); got != "" {
		t.Fatalf("DefaultPath(empty) = %q, want empty", got)
	}
	home := t.TempDir()
	wantPath := filepath.Join(home, ".ssh", "authorized_keys")
	if got := DefaultPath(home); got != wantPath {
		t.Fatalf("DefaultPath() = %q, want %q", got, wantPath)
	}

	writeFile(t, wantPath, "ssh-ed25519 AAAA test@host\n")
	idx, err := IndexDefault(home)
	if err != nil {
		t.Fatalf("IndexDefault(): %v", err)
	}
	if len(idx.Entries()) != 1 {
		t.Fatalf("IndexDefault() entries = %#v, want 1 entry", idx.Entries())
	}

	idx, err = IndexDefault(filepath.Join(home, "missing-home"))
	if err != nil {
		t.Fatalf("IndexDefault(missing): %v", err)
	}
	if len(idx.Entries()) != 0 {
		t.Fatalf("IndexDefault(missing) entries = %#v, want empty", idx.Entries())
	}
}
