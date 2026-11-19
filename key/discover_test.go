package key

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestDefaultDir(t *testing.T) {
	if got := DefaultDir(""); got != "" {
		t.Fatalf("DefaultDir(empty) = %q, want empty", got)
	}
	if got, want := DefaultDir("/home/alice"), filepath.Join("/home/alice", ".ssh"); got != want {
		t.Fatalf("DefaultDir() = %q, want %q", got, want)
	}
}

func TestDiscoverIdentityFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "id_ed25519"))
	writeFile(t, filepath.Join(dir, "id_ed25519.pub"))
	writeFile(t, filepath.Join(dir, "id_ed25519-cert.pub"))
	writeFile(t, filepath.Join(dir, "id_rsa"))
	writeFile(t, filepath.Join(dir, "known_hosts"))
	writeFile(t, filepath.Join(dir, "config"))
	if err := os.Mkdir(filepath.Join(dir, "id_dir"), 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}

	files, err := DiscoverIdentityFiles(dir)
	if err != nil {
		t.Fatalf("DiscoverIdentityFiles(): %v", err)
	}
	got := make([]string, 0, len(files))
	for _, f := range files {
		got = append(got, f.Name)
	}
	want := []string{"id_ed25519", "id_rsa"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DiscoverIdentityFiles() names = %#v, want %#v", got, want)
	}
	if files[0].Path != filepath.Join(dir, "id_ed25519") {
		t.Fatalf("DiscoverIdentityFiles()[0].Path = %q", files[0].Path)
	}
}

func TestDiscoverIdentityFilesMissingOrUnreadable(t *testing.T) {
	files, err := DiscoverIdentityFiles(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatalf("DiscoverIdentityFiles(missing): %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("DiscoverIdentityFiles(missing) = %#v, want empty", files)
	}

	if files, err := DiscoverIdentityFiles(""); err != nil || len(files) != 0 {
		t.Fatalf("DiscoverIdentityFiles(empty) = %#v, %v, want empty,nil", files, err)
	}
}

func TestDiscoverDefaultIdentityFiles(t *testing.T) {
	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	writeFile(t, filepath.Join(sshDir, "id_test"))

	files, err := DiscoverDefaultIdentityFiles(home)
	if err != nil {
		t.Fatalf("DiscoverDefaultIdentityFiles(): %v", err)
	}
	if len(files) != 1 || files[0].Name != "id_test" {
		t.Fatalf("DiscoverDefaultIdentityFiles() = %#v, want id_test", files)
	}
	if got, want := Paths(files), []string{filepath.Join(sshDir, "id_test")}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Paths() = %#v, want %#v", got, want)
	}
	if got := Paths(nil); len(got) != 0 {
		t.Fatalf("Paths(nil) = %#v, want empty", got)
	}
}

func TestKeyPaths(t *testing.T) {
	if got, want := PublicKeyPath("/home/alice/.ssh/id_ed25519"), "/home/alice/.ssh/id_ed25519.pub"; got != want {
		t.Fatalf("PublicKeyPath() = %q, want %q", got, want)
	}
	if got, want := CertificatePath("/home/alice/.ssh/id_ed25519"), "/home/alice/.ssh/id_ed25519-cert.pub"; got != want {
		t.Fatalf("CertificatePath() = %q, want %q", got, want)
	}
	if PublicKeyPath("") != "" || CertificatePath("") != "" {
		t.Fatal("empty private key path should return empty derived paths")
	}
}

func TestPairIdentityFiles(t *testing.T) {
	paths := []string{
		"/home/alice/.ssh/id_ed25519",
		"/home/alice/.ssh/id_ed25519.pub",
		"/home/alice/.ssh/id_ed25519-cert.pub",
		"/home/alice/.ssh/id_rsa.pub",
		"/home/alice/.ssh/id_rsa",
		"/home/alice/.ssh/other-cert.pub",
		"",
	}
	pairs := PairIdentityFiles(paths)
	want := []KeyPair{
		{Private: "/home/alice/.ssh/id_ed25519", Public: "/home/alice/.ssh/id_ed25519.pub", Certificate: "/home/alice/.ssh/id_ed25519-cert.pub"},
		{Private: "/home/alice/.ssh/id_rsa", Public: "/home/alice/.ssh/id_rsa.pub"},
		{Certificate: "/home/alice/.ssh/other-cert.pub"},
	}
	if !reflect.DeepEqual(pairs, want) {
		t.Fatalf("PairIdentityFiles() = %#v, want %#v", pairs, want)
	}
}

func TestDiscoverKeyPairs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "id_ed25519"))
	writeFile(t, filepath.Join(dir, "id_ed25519.pub"))
	writeFile(t, filepath.Join(dir, "id_ed25519-cert.pub"))
	writeFile(t, filepath.Join(dir, "id_rsa"))
	writeFile(t, filepath.Join(dir, "known_hosts"))
	if err := os.Mkdir(filepath.Join(dir, "id_dir"), 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}

	pairs, err := DiscoverKeyPairs(dir)
	if err != nil {
		t.Fatalf("DiscoverKeyPairs(): %v", err)
	}
	want := []KeyPair{
		{Private: filepath.Join(dir, "id_ed25519"), Public: filepath.Join(dir, "id_ed25519.pub"), Certificate: filepath.Join(dir, "id_ed25519-cert.pub")},
		{Private: filepath.Join(dir, "id_rsa")},
	}
	if !reflect.DeepEqual(pairs, want) {
		t.Fatalf("DiscoverKeyPairs() = %#v, want %#v", pairs, want)
	}

	if pairs, err := DiscoverKeyPairs(filepath.Join(dir, "missing")); err != nil || len(pairs) != 0 {
		t.Fatalf("DiscoverKeyPairs(missing) = %#v, %v, want empty,nil", pairs, err)
	}
	if pairs, err := DiscoverKeyPairs(""); err != nil || len(pairs) != 0 {
		t.Fatalf("DiscoverKeyPairs(empty) = %#v, %v, want empty,nil", pairs, err)
	}

	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	writeFile(t, filepath.Join(sshDir, "id_test"))
	writeFile(t, filepath.Join(sshDir, "id_test.pub"))
	pairs, err = DiscoverDefaultKeyPairs(home)
	if err != nil {
		t.Fatalf("DiscoverDefaultKeyPairs(): %v", err)
	}
	want = []KeyPair{{Private: filepath.Join(sshDir, "id_test"), Public: filepath.Join(sshDir, "id_test.pub")}}
	if !reflect.DeepEqual(pairs, want) {
		t.Fatalf("DiscoverDefaultKeyPairs() = %#v, want %#v", pairs, want)
	}
}
