package knownhosts

import (
	"strings"
	"testing"
)

func TestUnknownMarkerLinesAreSkipped(t *testing.T) {
	content := strings.Join([]string{
		"@cert-authority ok-host ssh-rsa AAAA",
		"@cert-authorit typo-host ssh-rsa AAAA",
		"@weak future-host ssh-rsa AAAA",
		"@revoked revoked-host ssh-rsa AAAA",
		"plain-host ssh-rsa AAAA",
	}, "\n")
	idx, err := IndexReader(strings.NewReader(content))
	if err != nil {
		t.Fatalf("IndexReader(): %v", err)
	}
	got := idx.HostCandidates()
	want := []string{"ok-host", "plain-host"}
	if len(got) != len(want) {
		t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("HostCandidates() = %#v, want %#v", got, want)
		}
	}
}

func TestCompletableHostRejectsMarkerAndColonTokens(t *testing.T) {
	if isCompletableHost("@foo") {
		t.Fatal("@foo should not be completable")
	}
	var got []string
	parseLine("example.com:2222 ssh-rsa AAAA", func(h string) { got = append(got, h) })
	if len(got) != 0 {
		t.Fatalf("unbracketed host:port produced candidates %v", got)
	}
	var got6 []string
	parseLine("[2001:db8::1]:22 ssh-rsa AAAA", func(h string) { got6 = append(got6, h) })
	if len(got6) != 1 || got6[0] != "2001:db8::1" {
		t.Fatalf("bracketed IPv6 candidates = %v, want [2001:db8::1]", got6)
	}
}
