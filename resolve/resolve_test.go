package resolve

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestHostValidation(t *testing.T) {
	if _, err := Host("   "); err == nil {
		t.Fatal("Host(empty) error = nil, want error")
	}
}

func TestHostParsesOutput(t *testing.T) {
	runner := func(_ context.Context, sshPath string, args []string) ([]byte, error) {
		if sshPath != "ssh-custom" {
			t.Fatalf("sshPath = %q, want ssh-custom", sshPath)
		}
		wantArgs := []string{"-F", "/tmp/ssh_config", "-G", "--", "example.com"}
		if !reflect.DeepEqual(args, wantArgs) {
			t.Fatalf("args = %#v, want %#v", args, wantArgs)
		}
		return []byte(strings.Join([]string{
			"user alice",
			"hostname example.internal",
			"port 2222",
			"identityfile ~/.ssh/id_ed25519",
			"identityfile ~/.ssh/id_rsa",
			"proxyjump bastion.example.com",
			"userknownhostsfile ~/.ssh/known_hosts ~/.ssh/known_hosts2",
			"", // ignored
		}, "\n")), nil
	}

	res, err := Host("example.com",
		WithSSHPath("ssh-custom"),
		WithConfigPath("/tmp/ssh_config"),
		WithRunner(runner),
	)
	if err != nil {
		t.Fatalf("Host(): %v", err)
	}
	if res.Host != "example.com" {
		t.Fatalf("Host = %q, want example.com", res.Host)
	}
	if got, want := res.First("user"), "alice"; got != want {
		t.Fatalf("First(user) = %q, want %q", got, want)
	}
	if got, want := res.First("HostName"), "example.internal"; got != want {
		t.Fatalf("First(HostName) = %q, want %q", got, want)
	}
	if got, want := res.All("identityfile"), []string{"~/.ssh/id_ed25519", "~/.ssh/id_rsa"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("All(identityfile) = %#v, want %#v", got, want)
	}
	if got := res.All("missing"); got != nil {
		t.Fatalf("All(missing) = %#v, want nil", got)
	}
	if got, want := res.Keys(), []string{"user", "hostname", "port", "identityfile", "proxyjump", "userknownhostsfile"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys() = %#v, want %#v", got, want)
	}
}

func TestHostRunnerError(t *testing.T) {
	wantErr := errors.New("boom")
	_, err := Host("example.com", WithRunner(func(context.Context, string, []string) ([]byte, error) {
		return nil, wantErr
	}))
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("Host() err = %v, want wrapped boom", err)
	}
}

func TestHostContext(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "ok")
	_, err := HostContext(ctx, "example", WithRunner(func(got context.Context, _ string, _ []string) ([]byte, error) {
		if got.Value(ctxKey{}) != "ok" {
			t.Fatalf("context value missing")
		}
		return []byte("hostname example\n"), nil
	}))
	if err != nil {
		t.Fatalf("HostContext(): %v", err)
	}

	_, err = HostContext(nil, "example", WithRunner(func(got context.Context, _ string, _ []string) ([]byte, error) {
		if got == nil {
			t.Fatal("ctx should not be nil")
		}
		return []byte("hostname example\n"), nil
	}))
	if err != nil {
		t.Fatalf("HostContext(nil ctx): %v", err)
	}
}

func TestNilResult(t *testing.T) {
	var r *Result
	if r.First("user") != "" {
		t.Fatal("nil First(user) should be empty")
	}
	if r.All("user") != nil {
		t.Fatal("nil All(user) should be nil")
	}
	if r.Keys() != nil {
		t.Fatal("nil Keys() should be nil")
	}
}
