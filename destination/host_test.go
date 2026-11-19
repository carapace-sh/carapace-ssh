package destination

import (
	"reflect"
	"testing"
)

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
		in     string
		user   string
		host   string
		port   string
		out    string
		isZero bool
	}{
		{"alice@example.com", "alice", "example.com", "", "alice@example.com", false},
		{"example.com:2222", "", "example.com", "2222", "example.com:2222", false},
		{"alice@[2001:db8::1]:22", "alice", "2001:db8::1", "22", "alice@[2001:db8::1]:22", false},
		{"alice@example.com:22", "alice", "example.com", "22", "alice@example.com:22", false},
		{"2001:db8::1", "", "2001:db8::1", "", "[2001:db8::1]", false},
		{"[2001:db8::1]", "", "2001:db8::1", "", "[2001:db8::1]", false},
		{"", "", "", "", "", true},
	}
	for _, tt := range tests {
		got := ParseTarget(tt.in)
		if got.User != tt.user || got.Host != tt.host || got.Port != tt.port {
			t.Fatalf("ParseTarget(%q) = %#v, want user=%q host=%q port=%q", tt.in, got, tt.user, tt.host, tt.port)
		}
		if got.IsZero() != tt.isZero {
			t.Fatalf("ParseTarget(%q).IsZero() = %v, want %v", tt.in, got.IsZero(), tt.isZero)
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

func TestJoinSCPRemote(t *testing.T) {
	tests := []struct {
		target string
		path   string
		want   string
	}{
		{"host", "/path", "host:/path"},
		{"user@[2001:db8::1]", "file.txt", "user@[2001:db8::1]:file.txt"},
		{"", "/local/path", "/local/path"},
		{"host", "", "host:"},
	}
	for _, tt := range tests {
		if got := JoinSCPRemote(tt.target, tt.path); got != tt.want {
			t.Fatalf("JoinSCPRemote(%q,%q) = %q, want %q", tt.target, tt.path, got, tt.want)
		}
	}
}

func TestProxyJump(t *testing.T) {
	tests := []struct {
		in   string
		hops []Target
		out  string
	}{
		{"", nil, ""},
		{"jump.example", []Target{{Host: "jump.example"}}, "jump.example"},
		{"user@jump.example:2222,bastion", []Target{{User: "user", Host: "jump.example", Port: "2222"}, {Host: "bastion"}}, "user@jump.example:2222,bastion"},
		{"alice@[2001:db8::1]:22, bob@host2", []Target{{User: "alice", Host: "2001:db8::1", Port: "22"}, {User: "bob", Host: "host2"}}, "alice@[2001:db8::1]:22,bob@host2"},
		{" , host1,,host2 , ", []Target{{Host: "host1"}, {Host: "host2"}}, "host1,host2"},
	}

	for _, tt := range tests {
		got := ParseProxyJump(tt.in)
		if len(got) != len(tt.hops) {
			t.Fatalf("ParseProxyJump(%q) = %#v, want %#v", tt.in, got, tt.hops)
		}
		for i := range got {
			if got[i] != tt.hops[i] {
				t.Fatalf("ParseProxyJump(%q)[%d] = %#v, want %#v", tt.in, i, got[i], tt.hops[i])
			}
		}
		if out := JoinProxyJump(got); out != tt.out {
			t.Fatalf("JoinProxyJump(ParseProxyJump(%q)) = %q, want %q", tt.in, out, tt.out)
		}
	}

	if got := SplitProxyJump("a,b"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("SplitProxyJump(%q) = %#v, want [a b]", "a,b", got)
	}
}

func TestIsURIAndIsRemote(t *testing.T) {
	if !IsURI("sftp://host/path") {
		t.Fatal("IsURI(sftp://host/path)=false, want true")
	}
	if IsURI("host:/path") {
		t.Fatal("IsURI(host:/path)=true, want false")
	}
	if !IsRemote("host:/path") {
		t.Fatal("IsRemote(host:/path)=false, want true")
	}
	if !IsRemote("sftp://host/path") {
		t.Fatal("IsRemote(sftp://host/path)=false, want true")
	}
	if IsRemote("/local/path") {
		t.Fatal("IsRemote(/local/path)=true, want false")
	}
}

func TestParseAndFormatURI(t *testing.T) {
	tests := []struct {
		in  string
		uri URI
		ok  bool
		out string
	}{
		{"sftp://example.com/path/to/file", URI{Scheme: "sftp", Host: "example.com", Path: "/path/to/file"}, true, "sftp://example.com/path/to/file"},
		{"sftp://alice@example.com:2222/dir", URI{Scheme: "sftp", User: "alice", Host: "example.com", Port: "2222", Path: "/dir"}, true, "sftp://alice@example.com:2222/dir"},
		{"scp://bob@[2001:db8::1]/x", URI{Scheme: "scp", User: "bob", Host: "2001:db8::1", Path: "/x"}, true, "scp://bob@[2001:db8::1]/x"},
		{"sftp://host/path%20with%20space?download=1#frag", URI{Scheme: "sftp", Host: "host", Path: "/path%20with%20space", RawQuery: "download=1", Fragment: "frag"}, true, "sftp://host/path%20with%20space?download=1#frag"},
		{"sftp://host?x=1", URI{Scheme: "sftp", Host: "host", Path: "", RawQuery: "x=1"}, true, "sftp://host?x=1"},
		{"sftp://host/#frag", URI{Scheme: "sftp", Host: "host", Path: "/", Fragment: "frag"}, true, "sftp://host/#frag"},
		{"SFTP://host", URI{Scheme: "sftp", Host: "host", Path: ""}, true, "sftp://host"},
		{"host:/path", URI{}, false, ""},
		{"not-a-uri", URI{}, false, ""},
		{"", URI{}, false, ""},
	}

	for _, tt := range tests {
		got, ok := ParseURI(tt.in)
		if ok != tt.ok {
			t.Fatalf("ParseURI(%q) ok=%v, want %v", tt.in, ok, tt.ok)
		}
		if ok && !reflect.DeepEqual(got, tt.uri) {
			t.Fatalf("ParseURI(%q) = %#v, want %#v", tt.in, got, tt.uri)
		}
		if ok {
			if out := FormatURI(got); out != tt.out {
				t.Fatalf("FormatURI(ParseURI(%q)) = %q, want %q", tt.in, out, tt.out)
			}
		}
	}

	if out := FormatURI(URI{Scheme: "sftp", Host: "example.com", Path: "relative"}); out != "sftp://example.com/relative" {
		t.Fatalf("FormatURI(relative path) = %q, want %q", out, "sftp://example.com/relative")
	}
	if out := FormatURI(URI{Host: "example.com"}); out != "" {
		t.Fatalf("FormatURI(missing scheme) = %q, want empty", out)
	}
}

func TestParseRemote(t *testing.T) {
	tests := []struct {
		in     string
		remote Remote
		ok     bool
		out    string
	}{
		{"sftp://alice@example.com:2222/dir", Remote{Scheme: "sftp", Target: Target{User: "alice", Host: "example.com", Port: "2222"}, Path: "/dir", URI: true}, true, "sftp://alice@example.com:2222/dir"},
		{"scp://host/path", Remote{Scheme: "scp", Target: Target{Host: "host"}, Path: "/path", URI: true}, true, "scp://host/path"},
		{"sftp://host/path%20with%20space?download=1#frag", Remote{Scheme: "sftp", Target: Target{Host: "host"}, Path: "/path%20with%20space", RawQuery: "download=1", Fragment: "frag", URI: true}, true, "sftp://host/path%20with%20space?download=1#frag"},
		{"sftp://host?x=1", Remote{Scheme: "sftp", Target: Target{Host: "host"}, Path: "", RawQuery: "x=1", URI: true}, true, "sftp://host?x=1"},
		{"user@host:/path", Remote{Target: Target{User: "user", Host: "host"}, Path: "/path"}, true, "user@host:/path"},
		{"[2001:db8::1]:/file", Remote{Target: Target{Host: "2001:db8::1"}, Path: "/file"}, true, "[2001:db8::1]:/file"},
		{"host:", Remote{Target: Target{Host: "host"}, Path: ""}, true, "host:"},
		{":path", Remote{Path: "path"}, true, ":path"},
		{"/local/path", Remote{}, false, ""},
		{"relative", Remote{}, false, ""},
		{"C:\\path", Remote{}, false, ""},
		{"", Remote{}, false, ""},
	}

	for _, tt := range tests {
		got, ok := ParseRemote(tt.in)
		if ok != tt.ok {
			t.Fatalf("ParseRemote(%q) ok=%v, want %v (got=%#v)", tt.in, ok, tt.ok, got)
		}
		if !ok {
			continue
		}
		if got != tt.remote {
			t.Fatalf("ParseRemote(%q) = %#v, want %#v", tt.in, got, tt.remote)
		}
		if out := FormatRemote(got); out != tt.out {
			t.Fatalf("FormatRemote(ParseRemote(%q)) = %q, want %q", tt.in, out, tt.out)
		}
	}
}
