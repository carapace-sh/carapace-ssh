package option

import "testing"

func TestLookupMatchCriterion(t *testing.T) {
	tests := []struct {
		name    string
		kind    MatchArgKind
		hosts   bool
		known   bool
		zeroArg bool
	}{
		{name: "host", kind: MatchArgValue, hosts: true, known: true},
		{name: "!originalhost", kind: MatchArgValue, hosts: true, known: true},
		{name: "exec", kind: MatchArgRest, known: true},
		{name: "command=echo hi", kind: MatchArgRest, known: true},
		{name: "all", kind: MatchArgNone, known: true, zeroArg: true},
		{name: "canonical", kind: MatchArgNone, known: true, zeroArg: true},
		{name: "final", kind: MatchArgNone, known: true, zeroArg: true},
		{name: "user", kind: MatchArgValue, known: true},
		{name: "unknowncriterion", known: false},
	}

	for _, tt := range tests {
		c, ok := LookupMatchCriterion(tt.name)
		if ok != tt.known {
			t.Fatalf("LookupMatchCriterion(%q) known=%v, want %v", tt.name, ok, tt.known)
		}
		if !ok {
			continue
		}
		if c.Kind != tt.kind {
			t.Fatalf("LookupMatchCriterion(%q).Kind=%v, want %v", tt.name, c.Kind, tt.kind)
		}
		if c.ExtractsHosts != tt.hosts {
			t.Fatalf("LookupMatchCriterion(%q).ExtractsHosts=%v, want %v", tt.name, c.ExtractsHosts, tt.hosts)
		}
		if got := IsZeroArgMatchCriterion(tt.name); got != tt.zeroArg {
			t.Fatalf("IsZeroArgMatchCriterion(%q)=%v, want %v", tt.name, got, tt.zeroArg)
		}
	}
}
