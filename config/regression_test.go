package config

import (
	"reflect"
	"testing"
)

func TestSplitArgsKeepsUTF8Intact(t *testing.T) {
	tests := []struct {
		name string
		line string
		args []string
	}{
		{
			name: "cyrillic token with continuation byte 0xA0",
			line: "Match host se\u0420ver",
			args: []string{"host", "se\u0420ver"},
		},
		{
			name: "token containing NBSP",
			line: "Host h\u00a0ost",
			args: []string{"h\u00a0ost"},
		},
		{
			name: "ascii tokens still split",
			line: "Host a b\tc",
			args: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := parseDirectives([]byte(tt.line + "\n"))
			if len(d) != 1 {
				t.Fatalf("directives len = %d, want 1", len(d))
			}
			if d[0].keyword != "match" && d[0].keyword != "host" {
				t.Fatalf("keyword = %q", d[0].keyword)
			}
			if !reflect.DeepEqual(d[0].args, tt.args) {
				t.Fatalf("args = %#v, want %#v", d[0].args, tt.args)
			}
		})
	}
}

func TestMatchCriterionValueDoesNotSwallowKnownCriterion(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "inline empty value followed by criterion", args: []string{"host=", "user", "bob"}, want: nil},
		{name: "spaced equals followed by criterion", args: []string{"host", "=", "user", "bob"}, want: nil},
		{name: "inline empty followed by non-criterion", args: []string{"host=", "value"}, want: []string{"value"}},
		{name: "inline non-empty value", args: []string{"host=pattern", "user"}, want: []string{"pattern"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractMatchHosts(tt.args)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("extractMatchHosts(%#v) = %#v, want %#v", tt.args, got, tt.want)
			}
		})
	}
}
