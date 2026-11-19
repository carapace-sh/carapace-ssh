package authorizedkeys

import (
	"strings"
	"testing"
)

func FuzzParseAuthorizedKeys(f *testing.F) {
	seeds := []string{
		"",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test@example.com\n",
		"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQ== user@host\n",
		"# comment\n\n",
		`command="echo hello",no-port-forwarding ssh-ed25519 AAAA test`,
		`from="*.example.com,!bad.example.com" ssh-rsa AAAA`,
		`environment="FOO=bar",pty ssh-ed25519 AAAA env`,
		"ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY= ecdsa",
		"sk-ssh-ed25519@openssh.com AAAA sk",
		"ssh-ed25519-cert-v01@openssh.com AAAA cert",
		"invalid line without keytype",
		"ssh-ed25519",
		"restrict,command=\"/bin/false\" ssh-rsa AAAA limited",
		"\x00\xff\n",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		_, _ = IndexReader(strings.NewReader(data))
		_, _ = ParseLine(data)
	})
}
