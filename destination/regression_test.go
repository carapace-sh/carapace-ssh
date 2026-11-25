package destination

import "testing"

func TestSplitSCPRemoteSingleLetterHost(t *testing.T) {
	target, path, ok := SplitSCPRemote("a:file")
	if !ok || target != "a" || path != "file" {
		t.Fatalf("SplitSCPRemote(\"a:file\") = (%q, %q, %v), want (\"a\", \"file\", true)", target, path, ok)
	}
}

func TestIsWindowsDrivePath(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"C:\\path", true},
		{"C:/path", true},
		{"C:", true},
		{"a:file", false},
		{"c:path", false},
		{"1:\\path", false},
	}
	for _, tt := range tests {
		if got := isWindowsDrivePath(tt.in); got != tt.want {
			t.Fatalf("isWindowsDrivePath(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
