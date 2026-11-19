package destination

import "testing"

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		in   string
		host string
		port string
	}{
		{"example.com", "example.com", ""},
		{"example.com:22", "example.com", "22"},
		{"[example.com]", "example.com", ""},
		{"[example.com]:2222", "example.com", "2222"},
		{"[2001:db8::1]:22", "2001:db8::1", "22"},
		{"2001:db8::1", "2001:db8::1", ""},
		{"", "", ""},
	}
	for _, tt := range tests {
		got := ParseHostPort(tt.in)
		if got.Host != tt.host || got.Port != tt.port {
			t.Fatalf("ParseHostPort(%q) = %#v, want host=%q port=%q", tt.in, got, tt.host, tt.port)
		}
	}
}

func TestFormatHostPort(t *testing.T) {
	tests := []struct {
		host string
		port string
		want string
	}{
		{"example.com", "", "example.com"},
		{"example.com", "22", "example.com:22"},
		{"2001:db8::1", "", "[2001:db8::1]"},
		{"2001:db8::1", "22", "[2001:db8::1]:22"},
		{"[2001:db8::1]", "22", "[2001:db8::1]:22"},
		{"", "22", ""},
	}
	for _, tt := range tests {
		if got := FormatHostPort(tt.host, tt.port); got != tt.want {
			t.Fatalf("FormatHostPort(%q,%q) = %q, want %q", tt.host, tt.port, got, tt.want)
		}
	}
}

func TestParseAndFormatTarget(t *testing.T) {
	tests := []struct {
		in   string
		user string
		host string
		port string
		out  string
	}{
		{"alice@example.com", "alice", "example.com", "", "alice@example.com"},
		{"example.com:2222", "", "example.com", "2222", "example.com:2222"},
		{"alice@[2001:db8::1]:22", "alice", "2001:db8::1", "22", "alice@[2001:db8::1]:22"},
		{"alice@example.com:22", "alice", "example.com", "22", "alice@example.com:22"},
		{"", "", "", "", ""},
	}
	for _, tt := range tests {
		got := ParseTarget(tt.in)
		if got.User != tt.user || got.Host != tt.host || got.Port != tt.port {
			t.Fatalf("ParseTarget(%q) = %#v, want user=%q host=%q port=%q", tt.in, got, tt.user, tt.host, tt.port)
		}
		if out := FormatTarget(got); out != tt.out {
			t.Fatalf("FormatTarget(ParseTarget(%q)) = %q, want %q", tt.in, out, tt.out)
		}
	}
}

func TestSplitSCPRemote(t *testing.T) {
	tests := []struct {
		in     string
		target string
		path   string
		ok     bool
	}{
		{"host:/path", "host", "/path", true},
		{"user@host:file.txt", "user@host", "file.txt", true},
		{"[2001:db8::1]:/path", "[2001:db8::1]", "/path", true},
		{":path", "", "path", true},
		{"/local/path", "", "/local/path", false},
		{"relative", "", "relative", false},
		{"C:\\path", "", "C:\\path", false},
		{"", "", "", false},
	}
	for _, tt := range tests {
		target, path, ok := SplitSCPRemote(tt.in)
		if target != tt.target || path != tt.path || ok != tt.ok {
			t.Fatalf("SplitSCPRemote(%q) = (%q,%q,%v), want (%q,%q,%v)", tt.in, target, path, ok, tt.target, tt.path, tt.ok)
		}
	}
}
