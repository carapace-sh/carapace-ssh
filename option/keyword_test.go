package option

import "testing"

func TestLookupKeyword(t *testing.T) {
	tests := []struct {
		name      string
		canonical string
		known     bool
	}{
		{name: "Host", canonical: "Host", known: true},
		{name: "hostname", canonical: "HostName", known: true},
		{name: "  userKnownHostsFile  ", canonical: "UserKnownHostsFile", known: true},
		{name: "UnknownKeyword", known: false},
	}

	for _, tt := range tests {
		k, ok := LookupKeyword(tt.name)
		if ok != tt.known {
			t.Fatalf("LookupKeyword(%q) known=%v, want %v", tt.name, ok, tt.known)
		}
		if !ok {
			continue
		}
		if k.Name != tt.canonical {
			t.Fatalf("LookupKeyword(%q).Name=%q, want %q", tt.name, k.Name, tt.canonical)
		}
		if !IsKnownKeyword(tt.name) {
			t.Fatalf("IsKnownKeyword(%q)=false, want true", tt.name)
		}
		if got := NormalizeKeyword(tt.name); got == "" {
			t.Fatalf("NormalizeKeyword(%q) returned empty string", tt.name)
		}
	}

	if IsKnownKeyword("NoSuchKeyword") {
		t.Fatal("IsKnownKeyword(NoSuchKeyword)=true, want false")
	}
}

func TestKeywordsList(t *testing.T) {
	keywords := Keywords()
	if len(keywords) == 0 {
		t.Fatal("Keywords() returned empty list")
	}
	for i := 1; i < len(keywords); i++ {
		if keywords[i-1].Name > keywords[i].Name {
			t.Fatalf("Keywords() not sorted: %#v", keywords)
		}
	}

	want := map[string]bool{
		"Host": true, "Match": true, "Include": true, "HostName": true, "User": true, "Port": true,
		"IdentityFile": true, "CertificateFile": true, "ProxyJump": true, "ProxyCommand": true,
		"UserKnownHostsFile": true, "GlobalKnownHostsFile": true,
	}
	for _, k := range keywords {
		delete(want, k.Name)
	}
	if len(want) != 0 {
		t.Fatalf("Keywords() missing %#v", want)
	}
}
