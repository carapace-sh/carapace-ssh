package knownhosts

import (
	"strings"
	"testing"
)

func FuzzIndexReader(f *testing.F) {
	seeds := []string{
		"",
		"github.com ssh-ed25519 AAAA...",
		"example.com,192.0.2.1 ssh-rsa AAAA...",
		"[example.com]:2222 ssh-rsa AAAA...",
		"|1|salt|hash ssh-rsa AAAA...",
		"@cert-authority *.example.com ssh-rsa AAAA...",
		"@revoked revoked.example.com ssh-rsa AAAA...",
		"# comment\n\n",
		"[2001:db8::1]:22 ssh-ed25519 AAAA...",
		"*.wildcard.example.com ssh-rsa AAAA...",
		"!negated.example.com ssh-rsa AAAA...",
		"\xff\xfe",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		_, _ = IndexReader(strings.NewReader(data))
	})
}
