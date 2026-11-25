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
		"@revoked revoked.example.com,still-revoked ssh-rsa AAAA...",
		"# comment\n\n",
		"[2001:db8::1]:22 ssh-ed25519 AAAA...",
		"[2001:db8::2] ssh-ed25519 AAAA...",
		"*.wildcard.example.com,good.example.com ssh-rsa AAAA...",
		"!negated.example.com,ok.example.com ssh-rsa AAAA...",
		"@cert-authority ca-only.example ssh-ed25519 AAAA...",
		"bad.example !negated,good.example ssh-ed25519 AAAA...",
		"[192.0.2.1]:2222 ssh-ed25519 AAAA...",
		"host1.example,host2.example,*.wild,|1|abc|def ssh-rsa AAAA...",
		"@revoked [2001:db8::9]:22 ssh-ed25519 AAAA...",
		"\xff\xfe",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		_, _ = IndexReader(strings.NewReader(data))
	})
}
